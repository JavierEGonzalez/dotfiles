local map = vim.api.nvim_set_keymap


map('n', '<leader>w', ':w<cr>', { noremap = true, silent = false })
map('n', '<leader>ww', ':w<cr>', { noremap = true, silent = false })
map('n', '<leader>W', ':wall<cr>', { noremap = true, silent = false })
map('n', '<leader>wq', ':wq<cr>', { noremap = true, silent = false })
map('n', '<leader>qq', ':wq<cr>', { noremap = true, silent = false })
map('n', '<leader>Q', ':wqall!<cr>', { noremap = true, silent = false })
map('n', '<leader>QQ', ':wqall!<cr>', { noremap = true, silent = false })

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

-- toggle Ntree (Lexplore) 25 characters wide
map('n', '`', ':25Lexplore<cr>', { noremap = true, silent = false })

-- eslint language server has to be installed
map('n', '<leader>fa', ':EsLintFixAll<cr>', { noremap = true, silent = false })

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
    bind('o', 'R')
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

