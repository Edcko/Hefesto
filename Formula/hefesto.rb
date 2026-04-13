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
      sha256 "e1a7d9005aa3d9af664a6ebde46208e5b0edce0f109b5ec120b199f4c1e4fdf3"
    end
    on_intel do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-darwin-amd64"
      sha256 "abb1c2ba03036e8dcca4178a0911cbb21790e845992b8afbd365268c7d80be6f"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-linux-arm64"
      sha256 "88269fce348e0a3ac990814c6334e78eb3f4319d138bbbbf6cd87737c7e3e7f9"
    end
    on_intel do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-linux-amd64"
      sha256 "4caae872ca03d9c0d7a87cca5820a84b0ef1de5cdf3a3bc6cf78e65e498836e1"
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
