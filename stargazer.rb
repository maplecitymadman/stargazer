# Homebrew Formula for Stargazer
# To install: brew install --build-from-source stargazer.rb
# Or add to your tap: brew tap your-org/stargazer

class Stargazer < Formula
  desc "Kubernetes troubleshooting tool with AI-powered diagnostics"
  homepage "https://github.com/maplecitymadman/stargazer"
  url "https://github.com/maplecitymadman/stargazer/archive/refs/heads/main.tar.gz"
  version "0.1.0"
  sha256 "" # Update with actual SHA256 when releasing
  license "MIT"

  depends_on "go" => :build

  def install
    # Build the binary
    system "go", "build", "-ldflags", "-s -w -X main.version=#{version}",
           "-o", bin/"stargazer", "cmd/stargazer/main.go"

    # Create config directory
    (buildpath/".stargazer").mkpath

    # Generate man pages (if you have them)
    # system "go", "run", "cmd/docs/main.go", "--man", man1
  end

  test do
    # Test that the binary works
    system "#{bin}/stargazer", "--version"

    # Test that it can at least show help
    system "#{bin}/stargazer", "--help"
  end
end
