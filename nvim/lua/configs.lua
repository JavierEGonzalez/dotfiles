local map = vim.api.nvim_set_keymap


map('n', '<leader>w', ':w<cr>', { noremap = true, silent = false })
map('n', '<leader>ww', ':w<cr>', { noremap = true, silent = false })
map('n', '<leader>W', ':wall<cr>', { noremap = true, silent = false })
map('n', '<leader>wq', ':wq<cr>', { noremap = true, silent = false })
map('n', '<leader>qq', ':wq<cr>', { noremap = true, silent = false })
map('n', '<leader>Q', ':wqall!<cr>', { noremap = true, silent = false })
map('n', '<leader>QQ', ':wqall!<cr>', { noremap = true, silent = false })
map('n', '<leader>mp', '%', { noremap = true, silent = false })
map('n', '<leader><C-w>', ':close<cr>', { noremap = true, silent = false })
map('n', '>', '>>', { noremap = true, silent = false })
map('n', '<', '<<', { noremap = true, silent = false })

-- toggle Ntree (Lexplore) 25 characters wide
map('n', '`', ':25Lexplore<cr>', { noremap = true, silent = false })

-- common typing mistakes
vim.api.nvim_command(':command WQ wq')
vim.api.nvim_command(':command Wq wq')

vim.api.nvim_create_autocmd('filetype', {
  pattern = 'netrw',
  desc = 'Better mappings for netrw',
  callback = function()
    local bind = function(lhs, rhs)
      vim.keymap.set('n', lhs, rhs, {remap = true, buffer = true})
    end 

    -- edit new file
    bind('n', '%')

    -- rename file
    bind('r', 'R')
  end
})

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

