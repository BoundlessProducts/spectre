class Spectre < Formula
  desc "Formal specification language and Rust verification toolchain"
  homepage "https://github.com/akkeshavan/spectre"
  url "https://github.com/akkeshavan/spectre.git",
      branch: "main"
  version "0.3.1"
  license "MIT"

  depends_on "go" => :build
  depends_on "rust" => :build
  depends_on "z3"

  def install
    # Build the Go CLI
    system "go", "build", "-trimpath", "-o", "spectre", "./cmd/spectre"
    bin.install "spectre"

    # Build the Rust AST miner using vendored dependencies (no network needed)
    cd "rust/spectre-mine-rs" do
      system "cargo", "build", "--release", "--offline"
      bin.install "target/release/spectre-mine-rs"
    end

    # Install examples
    (share/"spectre/examples").install Dir["examples/*.spec"]
    (share/"spectre/examples/rust").install Dir["examples/rust/*.rs"]
  end

  test do
    # Verify a simple counter spec
    (testpath/"counter.spec").write <<~SPEC
      var counter: int
      init { counter = 0 }
      action increment { counter' = counter + 1 }
      invariant nonNeg { counter >= 0 }
    SPEC
    system "#{bin}/spectre", "verify", "counter.spec"
  end
end
