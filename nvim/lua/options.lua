vim.opt.number = true
vim.opt.relativenumber = true
vim.opt.splitbelow = true

vim.opt.tabstop = 2
vim.opt.softtabstop = 2
vim.opt.shiftwidth = 2
vim.opt.expandtab = true

vim.opt.mouse = 'a'
vim.opt.showmode = false


vim.opt.backspace=indent,eol,start
-- vim.opt.paste = true

vim.opt.scrolloff=999

vim.opt.hlsearch = true
vim.opt.incsearch = true

-- shows search and replace at the bottom of screen
vim.opt.inccommand = "split"

vim.opt.smartcase = true
vim.opt.wrap = false

vim.opt.termguicolors = true

-- virtualedit allows you to move into spaces where 
-- there are no characters
vim.opt.virtualedit = "block"

-- kickstart
vim.g.mapleader = ' '
vim.g.maplocalleader = ' '

-- Save undo history
vim.opt.undofile = true

vim.g.have_nerd_font = true

-- set up grep as the default search tool and exclude common directories I want to ignore
vim.opt.grepprg = "grep -rn --exclude-dir={.nuxt,dist,.generated,node_modules,.git,.yarn,.output} --exclude=yarn.lock"
vim.opt.grepformat ='%f:%l:%m'
