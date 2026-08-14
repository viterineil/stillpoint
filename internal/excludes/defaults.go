package excludes

// NodeDependencies is deliberately narrow. A directory named "packages" is
// not excluded because it commonly contains monorepo source code.
var NodeDependencies = []string{
	"**/node_modules/**",
	"**/.npm/**",
	"**/.pnpm-store/**",
	"**/.cache/node-gyp/**",
}

func Defaults() []string {
	return append([]string(nil), NodeDependencies...)
}
