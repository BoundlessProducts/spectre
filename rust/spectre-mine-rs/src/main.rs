//! spectre-mine-rs: AST-based Rust source extractor for Spectre spec mining.
//!
//! Parses a .rs file with `syn`, walks the typed AST, and emits a JSON object
//! matching the MinedSpec schema consumed by `spectre mine --lang rust`.
//!
//! Usage:
//!   spectre-mine-rs <file.rs> [--spec-name <name>]
//!
//! Exit codes:
//!   0  success — JSON written to stdout
//!   1  usage / IO error
//!   2  Rust parse error (caller should fall back to regex miner)

use std::collections::{HashMap, HashSet};
use proc_macro2::TokenStream;
use quote::ToTokens;
use serde::Serialize;
use syn::{
    visit::Visit, BinOp, Expr, ExprAssign, ExprBinary, ExprField, ExprIf,
    ExprMacro, File, ImplItem, Item, Lit, Member, Pat,
    Stmt, StmtMacro, Type, UnOp,
};

// ── JSON output schema ────────────────────────────────────────────────────────

#[derive(Serialize, Default)]
struct MinedSpec {
    spec_name:   String,
    struct_name: String,
    source_file: String,
    enums:       Vec<MinedEnum>,
    fields:      Vec<MinedField>,
    methods:     Vec<MinedMethod>,
}

#[derive(Serialize)]
struct MinedEnum {
    name:     String,
    variants: Vec<String>,
}

#[derive(Serialize)]
struct MinedField {
    name:         String,
    rust_type:    String,
    spectre_type: String,
    default:      String,
}

#[derive(Serialize)]
struct MinedMethod {
    name:     String,
    mut_self: bool,
    params:   Vec<MinedParam>,
    requires: Vec<String>,
    assigns:  Vec<MinedAssign>,
    body_raw: String,
}

#[derive(Serialize)]
struct MinedParam {
    name:         String,
    spectre_type: String,
}

#[derive(Serialize)]
struct MinedAssign {
    field:    String,
    op:       String,
    rhs_rust: String,
}

// ── Entry point ───────────────────────────────────────────────────────────────

fn main() {
    let args: Vec<String> = std::env::args().collect();
    if args.len() < 2 {
        eprintln!("Usage: spectre-mine-rs <file.rs> [--spec-name <name>]");
        std::process::exit(1);
    }

    let filename = &args[1];
    let mut spec_name = String::new();
    let mut i = 2;
    while i < args.len() {
        if args[i] == "--spec-name" && i + 1 < args.len() {
            spec_name = args[i + 1].clone();
            i += 2;
        } else {
            i += 1;
        }
    }

    let source = match std::fs::read_to_string(filename) {
        Ok(s)  => s,
        Err(e) => { eprintln!("Error reading {}: {}", filename, e); std::process::exit(1); }
    };

    let ast = match syn::parse_file(&source) {
        Ok(a)  => a,
        Err(e) => { eprintln!("Parse error in {}: {}", filename, e); std::process::exit(2); }
    };

    let result = extract_spec(&ast, filename, &spec_name);
    println!("{}", serde_json::to_string_pretty(&result).unwrap());
}

// ── Top-level extraction ──────────────────────────────────────────────────────

fn extract_spec(file: &File, filename: &str, spec_name: &str) -> MinedSpec {
    let enums      = extract_enums(file);
    let enum_names: HashSet<String> = enums.iter().map(|e| e.name.clone()).collect();

    let (mut fields, struct_name) = extract_struct_fields(file, spec_name, &enum_names);
    let field_names: HashSet<String> = fields.iter().map(|f| f.name.clone()).collect();

    // Pull literal init values from the `new` / `default` constructor.
    let inits = extract_constructor_inits(file, &struct_name, &field_names);
    for f in &mut fields {
        if let Some(init) = inits.get(&f.name) {
            f.default = init.clone();
        }
        // Resolve "EnumType.Default" placeholders to first real variant.
        if f.default.ends_with(".Default") {
            let type_name = f.default.trim_end_matches(".Default");
            if let Some(e) = enums.iter().find(|e| e.name == type_name) {
                if let Some(first) = e.variants.first() {
                    f.default = format!("{}.{}", type_name, first);
                }
            }
        }
    }

    let final_name = if spec_name.is_empty() { struct_name.clone() } else { spec_name.to_string() };
    let methods    = extract_methods(file, &struct_name, &field_names, &enum_names);

    MinedSpec {
        spec_name:   final_name,
        struct_name,
        source_file: filename.to_string(),
        enums,
        fields,
        methods,
    }
}

// ── Enum extraction ───────────────────────────────────────────────────────────

fn extract_enums(file: &File) -> Vec<MinedEnum> {
    let mut enums = Vec::new();
    for item in &file.items {
        if let Item::Enum(e) = item {
            // Only unit variants map cleanly to Spectre enums.
            let variants: Vec<String> = e.variants.iter()
                .filter(|v| matches!(v.fields, syn::Fields::Unit))
                .map(|v| v.ident.to_string())
                .collect();
            if !variants.is_empty() {
                enums.push(MinedEnum { name: e.ident.to_string(), variants });
            }
        }
    }
    enums
}

// ── Struct field extraction ───────────────────────────────────────────────────

fn extract_struct_fields(
    file: &File,
    target_name: &str,
    enum_names: &HashSet<String>,
) -> (Vec<MinedField>, String) {
    let mut chosen: Option<&syn::ItemStruct> = None;
    for item in &file.items {
        if let Item::Struct(s) = item {
            let n = s.ident.to_string();
            if !target_name.is_empty() && n.to_lowercase() == target_name.to_lowercase() {
                chosen = Some(s);
                break;
            }
            if chosen.is_none() { chosen = Some(s); }
        }
    }

    let s = match chosen {
        Some(s) => s,
        None    => return (Vec::new(), target_name.to_string()),
    };

    let struct_name = s.ident.to_string();
    let mut fields  = Vec::new();

    if let syn::Fields::Named(named) = &s.fields {
        for field in &named.named {
            let name = field.ident.as_ref().unwrap().to_string();
            let rust_type_str = ty_to_string(&field.ty);
            if rust_type_str.contains("PhantomData") { continue; }
            let (spectre_type, default) = type_to_spectre(&field.ty, enum_names);
            fields.push(MinedField { name, rust_type: rust_type_str, spectre_type, default });
        }
    }
    (fields, struct_name)
}

// ── Constructor init extraction ───────────────────────────────────────────────

fn extract_constructor_inits(
    file: &File,
    struct_name: &str,
    field_names: &HashSet<String>,
) -> HashMap<String, String> {
    let mut inits = HashMap::new();
    for item in &file.items {
        if let Item::Impl(impl_block) = item {
            if !impl_matches(impl_block, struct_name) { continue; }
            for impl_item in &impl_block.items {
                if let ImplItem::Fn(method) = impl_item {
                    let name = method.sig.ident.to_string();
                    if name != "new" && name != "default" { continue; }
                    collect_struct_literal_inits(&method.block.stmts, struct_name, field_names, &mut inits);
                }
            }
        }
    }
    inits
}

fn collect_struct_literal_inits(
    stmts: &[Stmt],
    struct_name: &str,
    field_names: &HashSet<String>,
    inits: &mut HashMap<String, String>,
) {
    for stmt in stmts {
        let expr = match stmt {
            Stmt::Expr(e, _) => e,
            Stmt::Local(local) => {
                if let Some(init) = &local.init {
                    scan_expr_for_struct_literal(&init.expr, struct_name, field_names, inits);
                }
                continue;
            }
            _ => continue,
        };
        scan_expr_for_struct_literal(expr, struct_name, field_names, inits);
    }
}

fn scan_expr_for_struct_literal(
    expr: &Expr,
    struct_name: &str,
    field_names: &HashSet<String>,
    inits: &mut HashMap<String, String>,
) {
    match expr {
        Expr::Struct(s) => {
            let last = s.path.segments.last().map(|s| s.ident.to_string()).unwrap_or_default();
            if last.to_lowercase() == struct_name.to_lowercase() || last == "Self" {
                for fv in &s.fields {
                    if let Member::Named(ident) = &fv.member {
                        let name = ident.to_string();
                        if field_names.contains(&name) {
                            let val = expr_to_spectre(&fv.expr);
                            if !val.is_empty() && is_literal_value(&val) {
                                inits.insert(name, val);
                            }
                        }
                    }
                }
            }
        }
        Expr::Block(b) => {
            collect_struct_literal_inits(&b.block.stmts, struct_name, field_names, inits);
        }
        _ => {}
    }
}

/// Returns true if s is a concrete literal (number, bool, string, enum variant, empty collection).
fn is_literal_value(s: &str) -> bool {
    s.parse::<f64>().is_ok()
        || s == "true"
        || s == "false"
        || s.starts_with('"')
        || (s.contains('.') && !s.contains(' ')) // enum variant: Type.Variant
        || s == "{}"
}

// ── Method extraction ─────────────────────────────────────────────────────────

const SKIP_METHODS: &[&str] = &[
    "new", "default", "clone", "fmt", "drop", "from", "into",
    "as_ref", "as_mut", "deref", "deref_mut", "borrow", "borrow_mut",
];

fn extract_methods(
    file: &File,
    struct_name: &str,
    field_names: &HashSet<String>,
    enum_names: &HashSet<String>,
) -> Vec<MinedMethod> {
    let mut methods = Vec::new();

    for item in &file.items {
        if let Item::Impl(impl_block) = item {
            if !impl_matches(impl_block, struct_name) { continue; }

            for impl_item in &impl_block.items {
                if let ImplItem::Fn(method) = impl_item {
                    let name = method.sig.ident.to_string();
                    if SKIP_METHODS.contains(&name.as_str()) { continue; }

                    let mut mut_self = false;
                    let mut params   = Vec::new();

                    for input in &method.sig.inputs {
                        match input {
                            syn::FnArg::Receiver(r) => {
                                mut_self = r.mutability.is_some();
                            }
                            syn::FnArg::Typed(pt) => {
                                let pname        = pat_to_name(&pt.pat);
                                let (stype, _)   = type_to_spectre(&pt.ty, enum_names);
                                params.push(MinedParam { name: pname, spectre_type: stype });
                            }
                        }
                    }

                    let mut extractor = BodyExtractor::new(field_names.clone());
                    extractor.visit_block(&method.block);

                    methods.push(MinedMethod {
                        name,
                        mut_self,
                        params,
                        requires: extractor.requires,
                        assigns:  extractor.assigns,
                        body_raw: String::new(),
                    });
                }
            }
        }
    }
    methods
}

// ── Body visitor ──────────────────────────────────────────────────────────────

struct BodyExtractor {
    field_names:   HashSet<String>,
    requires:      Vec<String>,
    assigns:       Vec<MinedAssign>,
    seen_requires: HashSet<String>,
    seen_assigns:  HashSet<String>,
}

impl BodyExtractor {
    fn new(field_names: HashSet<String>) -> Self {
        Self {
            field_names,
            requires:      Vec::new(),
            assigns:       Vec::new(),
            seen_requires: HashSet::new(),
            seen_assigns:  HashSet::new(),
        }
    }

    fn add_require(&mut self, cond: String) {
        if !cond.is_empty() && !self.seen_requires.contains(&cond) {
            self.seen_requires.insert(cond.clone());
            self.requires.push(cond);
        }
    }

    fn add_assign(&mut self, field: String, op: String, rhs: String) {
        if rhs.is_empty() { return; }
        let key = format!("{}{}{}", field, op, rhs);
        if !self.seen_assigns.contains(&key) {
            self.seen_assigns.insert(key);
            self.assigns.push(MinedAssign { field, op, rhs_rust: rhs });
        }
    }

    // Shared logic for assert!/debug_assert!/assert_eq!/assert_ne! in any position.
    fn handle_macro_call(&mut self, path_name: &str, tokens: proc_macro2::TokenStream) {
        match path_name {
            "assert" | "debug_assert" => {
                if let Ok(expr) = syn::parse2::<Expr>(tokens.clone()) {
                    self.add_require(expr_to_spectre(&expr));
                } else if let Ok(args) = syn::parse2::<CommaSep>(tokens) {
                    if let Some(first) = args.0.first() {
                        self.add_require(expr_to_spectre(first));
                    }
                }
            }
            "assert_eq" | "assert_ne" => {
                if let Ok(args) = syn::parse2::<CommaSep>(tokens) {
                    let v: Vec<&Expr> = args.0.iter().take(2).collect();
                    if v.len() == 2 {
                        let l  = expr_to_spectre(v[0]);
                        let r  = expr_to_spectre(v[1]);
                        let op = if path_name == "assert_eq" { "=" } else { "!=" };
                        if !l.is_empty() && !r.is_empty() {
                            self.add_require(format!("{} {} {}", l, op, r));
                        }
                    }
                }
            }
            _ => {}
        }
    }
}

impl<'ast> Visit<'ast> for BodyExtractor {
    // assert!() in statement position: Stmt::Macro (the common case with semicolon)
    fn visit_stmt_macro(&mut self, mac: &'ast StmtMacro) {
        let path_name = mac.mac.path.segments.last()
            .map(|s| s.ident.to_string())
            .unwrap_or_default();
        self.handle_macro_call(&path_name, mac.mac.tokens.clone());
        syn::visit::visit_stmt_macro(self, mac);
    }

    // assert!() in expression position (rare but possible)
    fn visit_expr_macro(&mut self, mac: &'ast ExprMacro) {
        let path_name = mac.mac.path.segments.last()
            .map(|s| s.ident.to_string())
            .unwrap_or_default();
        self.handle_macro_call(&path_name, mac.mac.tokens.clone());
        syn::visit::visit_expr_macro(self, mac);
    }

    // if cond { return Err(..) / panic!(..) } → require !cond
    fn visit_expr_if(&mut self, expr_if: &'ast ExprIf) {
        if is_error_return_block(&expr_if.then_branch) {
            let cond    = expr_to_spectre(&expr_if.cond);
            let negated = negate_cond(&cond);
            self.add_require(negated);
        }
        syn::visit::visit_expr_if(self, expr_if);
    }

    // self.field = expr
    fn visit_expr_assign(&mut self, assign: &'ast ExprAssign) {
        if let Some(field) = self_field_name(&assign.left) {
            if self.field_names.contains(&field) {
                let rhs = expr_to_spectre(&assign.right);
                self.add_assign(field, "=".into(), rhs);
            }
        }
        syn::visit::visit_expr_assign(self, assign);
    }

    // self.field += / -= / *= / /= expr  (compound assignments are ExprBinary in syn 2)
    fn visit_expr_binary(&mut self, binary: &'ast ExprBinary) {
        let op = match &binary.op {
            BinOp::AddAssign(_) => Some("+="),
            BinOp::SubAssign(_) => Some("-="),
            BinOp::MulAssign(_) => Some("*="),
            BinOp::DivAssign(_) => Some("/="),
            BinOp::RemAssign(_) => Some("%="),
            _ => None,
        };
        if let Some(op_str) = op {
            if let Some(field) = self_field_name(&binary.left) {
                if self.field_names.contains(&field) {
                    let rhs = expr_to_spectre(&binary.right);
                    self.add_assign(field, op_str.into(), rhs);
                }
            }
        }
        syn::visit::visit_expr_binary(self, binary);
    }
}

// ── Expression → Spectre translation ─────────────────────────────────────────

fn expr_to_spectre(expr: &Expr) -> String {
    match expr {
        // self.field → field name only
        Expr::Field(f) => {
            if let Some(name) = self_field_name_inner(f) {
                return name;
            }
            let base   = expr_to_spectre(&f.base);
            let member = match &f.member {
                Member::Named(i)   => i.to_string(),
                Member::Unnamed(i) => i.index.to_string(),
            };
            if base.is_empty() { member } else { format!("{}.{}", base, member) }
        }

        // Binary: translate operator, recurse on operands
        Expr::Binary(b) => {
            let l = expr_to_spectre(&b.left);
            let r = expr_to_spectre(&b.right);
            let op = match &b.op {
                BinOp::Eq(_)  => "=",
                BinOp::Ne(_)  => "!=",
                BinOp::Lt(_)  => "<",
                BinOp::Le(_)  => "<=",
                BinOp::Gt(_)  => ">",
                BinOp::Ge(_)  => ">=",
                BinOp::And(_) => "&&",
                BinOp::Or(_)  => "||",
                BinOp::Add(_) => "+",
                BinOp::Sub(_) => "-",
                BinOp::Mul(_) => "*",
                BinOp::Div(_) => "/",
                BinOp::Rem(_) => "%",
                _ => return String::new(),
            };
            if l.is_empty() || r.is_empty() { return String::new(); }
            format!("{} {} {}", l, op, r)
        }

        // !expr → negation; -expr → unary minus
        Expr::Unary(u) => match &u.op {
            UnOp::Not(_) => negate_cond(&expr_to_spectre(&u.expr)),
            UnOp::Neg(_) => format!("-{}", expr_to_spectre(&u.expr)),
            _ => expr_to_spectre(&u.expr),
        },

        // (expr) → preserve parens
        Expr::Paren(p) => {
            let inner = expr_to_spectre(&p.expr);
            if inner.is_empty() { inner } else { format!("({})", inner) }
        }

        // SomeType::Variant → SomeType.Variant; plain ident → as-is
        Expr::Path(p) => {
            let segs: Vec<String> = p.path.segments.iter()
                .map(|s| s.ident.to_string())
                .collect();
            match segs.as_slice() {
                [one]    => one.clone(),
                [a, b]   => format!("{}.{}", a, b),
                _        => segs.join("."),
            }
        }

        // Literals
        Expr::Lit(l) => lit_to_spectre(&l.lit),

        // as-cast: drop the type, keep the value
        Expr::Cast(c) => expr_to_spectre(&c.expr),

        // &expr / &mut expr → strip reference
        Expr::Reference(r) => expr_to_spectre(&r.expr),

        // Common method calls on collections / Options
        Expr::MethodCall(mc) => {
            let receiver = expr_to_spectre(&mc.receiver);
            match mc.method.to_string().as_str() {
                "len"      | "count" => format!("{}.size()", receiver),
                "is_empty"           => format!("{}.size() = 0", receiver),
                "is_some"            => format!("!({} = 0)", receiver),
                "is_none"            => format!("{} = 0", receiver),
                _                    => String::new(),
            }
        }

        // Block with a single trailing expression
        Expr::Block(b) => {
            if let Some(Stmt::Expr(e, None)) = b.block.stmts.last() {
                return expr_to_spectre(e);
            }
            String::new()
        }

        _ => String::new(),
    }
}

fn lit_to_spectre(lit: &Lit) -> String {
    match lit {
        Lit::Int(i)   => i.base10_digits().replace('_', ""),
        Lit::Float(f) => f.base10_digits().replace('_', ""),
        Lit::Bool(b)  => b.value.to_string(),
        Lit::Str(s)   => format!("{:?}", s.value()),
        _             => String::new(),
    }
}

// ── Type → Spectre translation ────────────────────────────────────────────────

fn type_to_spectre(ty: &Type, enum_names: &HashSet<String>) -> (String, String) {
    match ty {
        Type::Path(tp) => {
            let last = match tp.path.segments.last() {
                Some(s) => s,
                None    => return ("int".into(), "0".into()),
            };
            let name = last.ident.to_string();
            match name.as_str() {
                "i8" | "i16" | "i32" | "i64" | "i128" |
                "u8" | "u16" | "u32" | "u64" | "u128" |
                "isize" | "usize" => ("int".into(), "0".into()),
                "f32" | "f64"     => ("float".into(), "0.0".into()),
                "bool"            => ("bool".into(), "false".into()),
                "String" | "str"  => ("str".into(), r#""""#.into()),
                "Vec"             => {
                    let inner = first_generic(last, enum_names);
                    (format!("List[{}]", inner), "{}".into())
                }
                "HashSet" | "BTreeSet" | "IndexSet" => {
                    let inner = first_generic(last, enum_names);
                    (format!("Set[{}]", inner), "{}".into())
                }
                "HashMap" | "BTreeMap" | "IndexMap" => {
                    let (k, v) = two_generics(last, enum_names);
                    (format!("Map[{}, {}]", k, v), "{}".into())
                }
                // Wrapper types — unwrap to inner type
                "Option" | "Arc" | "Mutex" | "RwLock" | "Rc" |
                "RefCell" | "Box" | "Cell" | "Cow" => {
                    if let syn::PathArguments::AngleBracketed(ab) = &last.arguments {
                        if let Some(syn::GenericArgument::Type(inner)) = ab.args.first() {
                            return type_to_spectre(inner, enum_names);
                        }
                    }
                    ("int".into(), "0".into())
                }
                other => {
                    if enum_names.contains(other) {
                        (other.to_string(), format!("{}.Default", other))
                    } else {
                        ("int".into(), "0".into())
                    }
                }
            }
        }
        Type::Reference(r)  => type_to_spectre(&r.elem, enum_names),
        Type::Slice(s)       => {
            let (inner, _) = type_to_spectre(&s.elem, enum_names);
            (format!("List[{}]", inner), "{}".into())
        }
        Type::Array(a)       => {
            let (inner, _) = type_to_spectre(&a.elem, enum_names);
            (format!("List[{}]", inner), "{}".into())
        }
        _ => ("int".into(), "0".into()),
    }
}

fn first_generic(seg: &syn::PathSegment, enum_names: &HashSet<String>) -> String {
    if let syn::PathArguments::AngleBracketed(ab) = &seg.arguments {
        if let Some(syn::GenericArgument::Type(t)) = ab.args.first() {
            return type_to_spectre(t, enum_names).0;
        }
    }
    "int".into()
}

fn two_generics(seg: &syn::PathSegment, enum_names: &HashSet<String>) -> (String, String) {
    if let syn::PathArguments::AngleBracketed(ab) = &seg.arguments {
        let types: Vec<_> = ab.args.iter().collect();
        if types.len() >= 2 {
            if let (syn::GenericArgument::Type(k), syn::GenericArgument::Type(v)) =
                (&types[0], &types[1])
            {
                return (type_to_spectre(k, enum_names).0, type_to_spectre(v, enum_names).0);
            }
        }
    }
    ("str".into(), "int".into())
}

// ── Helpers ───────────────────────────────────────────────────────────────────

/// Returns the field name if `expr` is `self.field`.
fn self_field_name(expr: &Expr) -> Option<String> {
    if let Expr::Field(f) = expr { self_field_name_inner(f) } else { None }
}

fn self_field_name_inner(f: &ExprField) -> Option<String> {
    if let Expr::Path(p) = f.base.as_ref() {
        if p.path.is_ident("self") {
            if let Member::Named(ident) = &f.member {
                return Some(ident.to_string());
            }
        }
    }
    None
}

/// Returns true if the block contains only an early exit (return/panic/bail).
fn is_error_return_block(block: &syn::Block) -> bool {
    if block.stmts.is_empty() { return false; }
    block.stmts.iter().any(|stmt| match stmt {
        Stmt::Expr(Expr::Return(_), _) => true,
        Stmt::Expr(Expr::Macro(m), _) => {
            let name = m.mac.path.segments.last()
                .map(|s| s.ident.to_string())
                .unwrap_or_default();
            matches!(name.as_str(), "panic" | "unreachable" | "todo" | "bail")
        }
        Stmt::Macro(m) => {
            let name = m.mac.path.segments.last()
                .map(|s| s.ident.to_string())
                .unwrap_or_default();
            matches!(name.as_str(), "panic" | "unreachable" | "todo" | "bail")
        }
        _ => false,
    })
}

/// Negate a Spectre condition string.
fn negate_cond(cond: &str) -> String {
    if cond.is_empty() { return String::new(); }
    if let Some(inner) = cond.strip_prefix('!') { return inner.to_string(); }
    for (a, b) in &[
        (" != ", " = "), (" = ", " != "),
        (" < ",  " >= "), (" >= ", " < "),
        (" > ",  " <= "), (" <= ", " > "),
    ] {
        if cond.contains(a) { return cond.replacen(a, b, 1); }
    }
    // Boolean flag or complex expression.
    if cond.chars().all(|c| c.is_alphanumeric() || c == '_' || c == '.') {
        format!("!{}", cond)
    } else {
        format!("!({})", cond)
    }
}

/// Extract the name from a simple `Pat::Ident` pattern.
fn pat_to_name(pat: &Pat) -> String {
    match pat {
        Pat::Ident(i) => i.ident.to_string(),
        _             => "_".to_string(),
    }
}

/// Convert a `syn::Type` back to a compact Rust type string (for rust_type field).
fn ty_to_string(ty: &Type) -> String {
    let mut ts = TokenStream::new();
    ty.to_tokens(&mut ts);
    ts.to_string().replace(' ', "")
}

/// Returns true if `impl_block.self_ty` matches `struct_name` (by last path segment).
fn impl_matches(impl_block: &syn::ItemImpl, struct_name: &str) -> bool {
    if struct_name.is_empty() { return true; }
    match impl_block.self_ty.as_ref() {
        Type::Path(tp) => {
            let last = tp.path.segments.last()
                .map(|s| s.ident.to_string())
                .unwrap_or_default();
            last.to_lowercase() == struct_name.to_lowercase()
        }
        _ => false,
    }
}

// ── Comma-separated expression list parser ────────────────────────────────────

struct CommaSep(syn::punctuated::Punctuated<Expr, syn::token::Comma>);

impl syn::parse::Parse for CommaSep {
    fn parse(input: syn::parse::ParseStream) -> syn::Result<Self> {
        Ok(CommaSep(syn::punctuated::Punctuated::parse_terminated(input)?))
    }
}
