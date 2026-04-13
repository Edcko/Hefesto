# typed: false
# frozen_string_literal: true

# Hefesto - TUI installer for OpenCode configuration
# https://github.com/Edcko/Hefesto

class Hefesto < Formula
  desc "TUI installer for OpenCode configuration"
  homepage "https://github.com/Edcko/Hefesto"
  version "1.0.0" # Updated automatically by CI/CD
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-darwin-arm64"
      sha256 "0b423e1241a81d83548cd20c22924b56a14e1b94b0a2231a33532ca3325a57b8"
    end
    on_intel do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-darwin-amd64"
      sha256 "edc3b8c243a7e10c33aa59b7e94737aaf5c3f407d07f39ca15ce9ee1a59b79a0"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-linux-arm64"
      sha256 "ff606610f35200dc9fa48dcd4a680948f2aaca05064a38541c44c77a21c6b2d7"
    end
    on_intel do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-linux-amd64"
      sha256 "af46119c9b7d493f1c83ee2ad171427991f1613b9b059f240130a15da2f65b1c"
    end
  end

  def install
    # Download pre-compiled binary
    bin.install "hefesto-#{OS.kernel_name.downcase}-#{Hardware::CPU.intel? ? "amd64" : "arm64"}" => "hefesto"
  end

  test do
    # Test that the binary runs and shows help
    assert_match "hefesto", shell_output("#{bin}/hefesto --help", 0)
  end
end
