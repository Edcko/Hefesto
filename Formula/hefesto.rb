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
      sha256 "c080421b791966bacd1cfdab10a6b0cc5d1ca3412d54a71e4d4866cda4a726f1"
    end
    on_intel do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-darwin-amd64"
      sha256 "032923c05b7484fd7144b9613cdabf78cb26c918f02f7ce4cad8cbadd6cca49f"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-linux-arm64"
      sha256 "18a7405472eead65cbf9624bdfc05da45ece5edca0cd7679fdf3485044dd62cf"
    end
    on_intel do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-linux-amd64"
      sha256 "527782a8fc4ceeadcb843287811c33fa122d9c1771fae419322a7bd0ffcf7bad"
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
