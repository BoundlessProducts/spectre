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

    # Build the Rust AST miner (required for `spectre mine --lang rust`)
    system "cargo", "build", "--release",
           "--manifest-path", "rust/spectre-mine-rs/Cargo.toml"
    bin.install "rust/spectre-mine-rs/target/release/spectre-mine-rs"

    # Install examples
    (share/"spectre/examples").install Dir["examples/*.spec"]
    (share/"spectre/examples/rust").install Dir["examples/rust/*.rs"]
  end

  test do
    system "#{bin}/spectre", "verify", "#{share}/spectre/examples/counter.spec"
  end
end
