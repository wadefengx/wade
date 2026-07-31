# frozen_string_literal: true

class Wade < Formula
  desc "All-in-one Node.js version & npm/yarn/pnpm registry manager"
  homepage "https://github.com/wadefengx/wade"
  version "0.2.0"
  license "MIT"

  BASE_URL = "https://github.com/wadefengx/wade/releases/download/v0.2.0/"

  on_macos do
    if Hardware::CPU.arm?
      url "#{BASE_URL}wade-darwin-arm64.tar.gz"
      sha256 "2777ba43fe161adaade3f84ccfbff528b63c162754bcd9673d21b6c886b3f1cb"
    else
      url "#{BASE_URL}wade-darwin-amd64.tar.gz"
      sha256 "939e02b6803bba5a2c32414febf017f54f19c491997e51c8775c26d552c45c05"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "#{BASE_URL}wade-linux-arm64.tar.gz"
      # TODO: fill from release
    else
      url "#{BASE_URL}wade-linux-amd64.tar.gz"
      sha256 "4956f29d86d6ea7c772b04f0cad91153d424f0e73f5bb5f9d42ebb7a2e2df357"
    end
  end

  def install
    bin.install "wade"
  end

  def post_install
    system "#{bin}/wade", "setup"
    ohai "Run 'echo \"export PATH=\\\"$HOME/.wade/shims:$PATH\\\"\" >> ~/.zshrc' to enable Node shims"
  end

  test do
    assert_match "wade", shell_output("#{bin}/wade --help")
  end
end
