class Spectre < Formula
  desc "A programmer-friendly specification language inspired by TLA+ and Quint"
  homepage "https://github.com/akkeshavan/spectre"
  url "https://github.com/akkeshavan/spectre.git",
      branch: "main"
  version "0.1.0"
  license "MIT"

  depends_on "go" => :build

  def install
    cd buildpath do
      system "go", "build", "-o", "spectre", "./cmd/spectre"
      bin.install "spectre"
    end
  end

  test do
    system "#{bin}/spectre", "--version"
  end
end

