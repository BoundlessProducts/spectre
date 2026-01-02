class Spectre < Formula
  desc "A programmer-friendly specification language inspired by TLA+ and Quint"
  homepage "https://github.com/spectre-lang/spectre"
  url "https://github.com/spectre-lang/spectre/archive/v0.1.0.tar.gz"
  sha256 ""
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", "spectre", "./cmd/spectre"
    bin.install "spectre"
  end

  test do
    system "#{bin}/spectre", "--version"
  end
end

