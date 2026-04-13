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
      sha256 "131f5957c756329755d140adf8d863157ca4a3685b1405cd096bfe8ca93962e0"
    end
    on_intel do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-darwin-amd64"
      sha256 "a30373031cc28782000515436b801ddb1444b3de2b7c2a7462e611655808f39d"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-linux-arm64"
      sha256 "61aef87435f85aefb0c48252cffa19fa817046a4c95dfd8f8e3a27f5d89db0ac"
    end
    on_intel do
      url "https://github.com/Edcko/Hefesto/releases/download/v#{version}/hefesto-linux-amd64"
      sha256 "6dcafdaeaf6975ef45de6f900d58e92751e1218e16dccbcc7ae83c3a831f8824"
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
