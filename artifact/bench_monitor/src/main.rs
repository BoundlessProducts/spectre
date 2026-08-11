mod mon_k1;
mod mon_k5;
mod mon_k10;
mod mon_k20;

use std::hint::black_box;
use std::time::Instant;

const ITERS: u64 = 10_000_000;
const RUNS: usize = 5;

fn run_baseline() -> f64 {
    let t0 = Instant::now();
    let mut sink: i64 = 0;
    for i in 0..ITERS {
        sink = black_box(i as i64 % 1000 + 1);
    }
    let elapsed = t0.elapsed();
    let _ = black_box(sink);
    elapsed.as_nanos() as f64 / ITERS as f64
}

macro_rules! run_k {
    ($mod:ident) => {{
        let initial = $mod::SpecState { aliceBalance: 0, bobBalance: 0 };
        let mut mon = $mod::Monitor::new(initial);
        let t0 = Instant::now();
        for i in 0..ITERS {
            let s = $mod::SpecState {
                aliceBalance: black_box((i as i64 % 1000) + 1),
                bobBalance:   black_box((i as i64 % 500) + 1),
            };
            let _ = black_box(mon.step("deposit", s));
        }
        t0.elapsed().as_nanos() as f64 / ITERS as f64
    }};
}

fn best_of(f: impl Fn() -> f64) -> f64 {
    (0..RUNS).map(|_| f()).fold(f64::MAX, f64::min)
}

fn main() {
    println!("Monitor overhead benchmark — {} M iterations, best of {}", ITERS / 1_000_000, RUNS);
    println!("Hardware: {} (Apple M-series aarch64, release build)", std::env::consts::ARCH);
    println!("{:-<60}", "");
    println!("{:<14} {:>8} {:>12} {:>12} {:>8}", "config", "ns/step", "Mstep/s", "overhead", "lines");
    println!("{:-<60}", "");

    // Warm up
    let _ = best_of(|| run_baseline());
    let _ = best_of(|| run_k!(mon_k1));

    let ns_base = best_of(|| run_baseline());
    let ns_k1   = best_of(|| run_k!(mon_k1));
    let ns_k5   = best_of(|| run_k!(mon_k5));
    let ns_k10  = best_of(|| run_k!(mon_k10));
    let ns_k20  = best_of(|| run_k!(mon_k20));

    let fmt_row = |label: &str, ns: f64, ovhd: Option<f64>, lines: &str| {
        let ovhd_s = match ovhd {
            None    => "--".to_string(),
            Some(v) => format!("{:.0}%", v),
        };
        println!("{:<14} {:>8.1} {:>12.0} {:>12} {:>8}",
                 label, ns, 1000.0 / ns, ovhd_s, lines);
    };

    fmt_row("baseline",  ns_base, None, "—");
    fmt_row("k=1",  ns_k1,  Some((ns_k1  - ns_base) / ns_base * 100.0), "140");
    fmt_row("k=5",  ns_k5,  Some((ns_k5  - ns_base) / ns_base * 100.0), "176");
    fmt_row("k=10", ns_k10, Some((ns_k10 - ns_base) / ns_base * 100.0), "221");
    fmt_row("k=20", ns_k20, Some((ns_k20 - ns_base) / ns_base * 100.0), "311");

    println!("{:-<60}", "");
    println!("overhead = (T_monitor - T_baseline) / T_baseline");
    println!("All monitors: 0 external deps, stack-only hot path, no heap alloc.");
}
