local map = vim.api.nvim_set_keymap


map('n', '<leader>w', ':w<cr>', { noremap = true, silent = false })
map('n', '<leader>ww', ':w<cr>', { noremap = true, silent = false })
map('n', '<leader>W', ':wall<cr>', { noremap = true, silent = false })
map('n', '<leader>wq', ':wq<cr>', { noremap = true, silent = false })
map('n', '<leader>qq', ':wq<cr>', { noremap = true, silent = false })
map('n', '<leader>Q', ':wqall!<cr>', { noremap = true, silent = false })
map('n', '<leader>QQ', ':wqall!<cr>', { noremap = true, silent = false })

vim.api.nvim_create_user_command('WQ', 'wq', {})
vim.api.nvim_create_user_command('Wq', 'wq', {})
vim.api.nvim_create_user_command('W', 'w', {})
vim.api.nvim_create_user_command('Qa', 'qa', {})
vim.api.nvim_create_user_command('Q', 'q', {})

-- go to [m]atching [p]arenthesis
map('n', '<leader>mp', '%', { noremap = true, silent = false })
map('v', '<leader>mp', '%', { noremap = true, silent = false })

-- go to [m]atching Bracket
map('n', 'M', '%', { noremap = true, silent = false })
map('v', 'M', '%', { noremap = true, silent = false })

map('n', '<leader><C-w>', ':close<cr>', { noremap = true, silent = false })
map('n', '>', '>>', { noremap = true, silent = false })
map('n', '<', '<<', { noremap = true, silent = false })
map('n', '<leader>y', '\"+y', { noremap = true, silent = false })
map('v', '<leader>y', '\"+y', { noremap = true, silent = false })
map('n', '<leader>Y', '\"+y', { noremap = true, silent = false })
map('v', '<leader>Y', '\"+y', { noremap = true, silent = false })

map('n', '<C-j>', ':Treewalker Down<CR>', { noremap = true })
map('n', '<C-k>', ':Treewalker Up<CR>', { noremap = true })
map('n', '<C-h>', ':Treewalker Left<CR>', { noremap = true })
map('n', '<C-l>', ':Treewalker Right<CR>', { noremap = true })

-- open diagnostics dialog box for full details
map('n', '<leader>dd', ':lua vim.diagnostic.goto_next() <CR>', { noremap = true })
map('n', '<leader>do', ':lua vim.diagnostic.open_float() <CR>', { noremap = true })
map('n', '<leader>dp', ':lua vim.diagnostic.goto_prev() <CR>', { noremap = true })

-- toggle Ntree (Lexplore) 25 characters wide
-- overriden by mini.files
-- map('n', '`', ':25Lexplore<cr>', { noremap = true, silent = false })

local triggers = {"."}

-- vim.api.nvim_create_autocmd("InsertCharPre", {
--  buffer = vim.api.nvim_get_current_buf(),
--  callback = function()
--    if vim.fn.pumvisible() == 1 or vim.fn.state("m") == "m" then
--      return
--    end
--    local char = vim.v.char
--    if vim.list_contains(triggers, char) then
--      local key = vim.keycode("<C-x><C-n>")
--      vim.api.nvim_feedkeys(key, "m", false)
--    end
--  end
--})

