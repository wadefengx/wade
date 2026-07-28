# frozen_string_literal: true

class Wade < Formula
  desc "All-in-one Node.js version & npm/yarn/pnpm registry manager"
  homepage "https://github.com/wadefengx/wade"
  version "${VERSION}"
  license "MIT"

  BASE_URL = "https://github.com/wadefengx/wade/releases/download/v${VERSION}"

  on_macos do
    if Hardware::CPU.arm?
      url "#{BASE_URL}/wade-darwin-arm64.tar.gz"
      sha256 "${SHA_DARWIN_ARM64}"
    else
      url "#{BASE_URL}/wade-darwin-amd64.tar.gz"
      sha256 "${SHA_DARWIN_AMD64}"
    end
  end

  on_linux do
    url "#{BASE_URL}/wade-linux-amd64.tar.gz"
    sha256 "${SHA_LINUX_AMD64}"
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
