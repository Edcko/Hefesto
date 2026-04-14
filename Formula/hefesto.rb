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
      sha256 "ae7886c6bb87333331096a3bcabcaba0e9b13a905a07f80e8b2383212d6c8686"
    end
    on_intel do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-darwin-amd64"
      sha256 "51a1fb3a1b8d43545b76b89c5bfc93f597ace632d5145899ba4f2a47234b914e"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-linux-arm64"
      sha256 "21b7f88d4225ece7a2e49c464d6064ed04dd15ef424d4ea9bbec7742fc9923e4"
    end
    on_intel do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-linux-amd64"
      sha256 "f2136a7750900d9ab9512b529d2d3acdda505c359cfe225635b1cd24999992bf"
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
