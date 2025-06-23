local M = {}

function M.hello_world()
	print("$ticket")
end

-- Map a command to the function
vim.api.nvim_command('command! HelloWorld lua require("nvim/plugins/adusa/init").hello_world()')

return M
