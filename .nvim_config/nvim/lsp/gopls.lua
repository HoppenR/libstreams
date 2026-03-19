local streamshower_home = os.getenv("LIBSTREAMS_HOME")

if not streamshower_home then return end

return {
	name = 'gopls',
	cmd = { 'gopls', 'serve' },
	filetypes = { 'go', 'gomod', 'gowork', 'gotmpl' },
	root_dir = streamshower_home,
	settings = {
		gopls = {
			analyses = {
				unusedparams = true,
				shadow = true,
			},
			staticcheck = true,
			gofumpt = true,
			completeUnimported = true,
			usePlaceholders = true,
			directoryFilters = { "-.git", "-.vscode", "-.idea", "-.bin", "-bin" },
		},
	},
	init_options = {
		usePlaceholders = true,
	},
}
