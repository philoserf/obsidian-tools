brew "go"
brew "golangci/tap/golangci-lint" # Go linter
brew "markdownlint-cli"
brew "prettier"
brew "go-task/tap/go-task" do
  post_install do
    system "go", "install", "mvdan.cc/gofumpt@latest"
    system "go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
  end
end
