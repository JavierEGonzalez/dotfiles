local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"

local lsp_plugin_table = require("lsp")

if not (vim.uv or vim.loop).fs_stat(lazypath) then
	local lazyrepo = "https://github.com/folke/lazy.nvim.git"
	local out = vim.fn.system({
		"git",
		"clone",
		"--filter=blob:none",
		"--branch=stable",
		lazyrepo,
		lazypath,
	})
	if vim.v.shell_error ~= 0 then
		error("Error cloning lazy.nvim:\n" .. out)
	end
end ---@diagnostic disable-next-line: undefined-field
vim.opt.rtp:prepend(lazypath)

require("lazy").setup({
	-- sleuth for normalizing tab width based on file
	"tpope/vim-sleuth",
	{
		"m4xshen/autoclose.nvim",
		config = function()
			require("autoclose").setup({
				keys = {
					["("] = { escape = false, close = true, pair = "()" },
					["["] = { escape = false, close = true, pair = "[]" },
					["{"] = { escape = false, close = true, pair = "{}" },

					[">"] = { escape = true, close = false, pair = "<>" },
					[")"] = { escape = true, close = false, pair = "()" },
					["]"] = { escape = true, close = false, pair = "[]" },
					["}"] = { escape = true, close = false, pair = "{}" },

					['"'] = { escape = true, close = true, pair = '""' },
					["'"] = { escape = true, close = true, pair = "''" },
					["`"] = { escape = true, close = true, pair = "``" },
				},
				options = {
					disabled_filetypes = { "text" },
					disable_when_touch = false,
					touch_regex = "[%w(%[{]",
					pair_spaces = true,
					auto_indent = true,
					disable_command_mode = false,
				},
			})
		end,
	},
	{
		"rebelot/kanagawa.nvim",
		init = function()
			vim.cmd.colorscheme("kanagawa-dragon")
		end,
		config = function()
			require("kanagawa").setup({
				transparent = true,
			})
		end,
	},
	{ -- Adds git related signs to the gutter, as well as utilities for managing changes
		"lewis6991/gitsigns.nvim",
		opts = {
			signs = {
				add = { text = "+" },
				change = { text = "~" },
				delete = { text = "_" },
				topdelete = { text = "‾" },
				changedelete = { text = "~" },
			},
		},
	},
	{
		-- Fuzzy Finder (files, lsp, etc)
		"nvim-telescope/telescope.nvim",
		event = "VimEnter",
		branch = "0.1.4",
		dependencies = {
			"nvim-lua/plenary.nvim",
			{ -- If encountering errors, see telescope-fzf-native README for installation instructions
				"nvim-telescope/telescope-fzf-native.nvim",

				-- `build` is used to run some command when the plugin is installed/updated.
				-- This is only run then, not every time Neovim starts up.
				build = "make",

				-- `cond` is a condition used to determine whether this plugin should be
				-- installed and loaded.
				cond = function()
					return vim.fn.executable("make") == 1
				end,
			},
			{ "nvim-telescope/telescope-ui-select.nvim" },

			-- Useful for getting pretty icons, but requires a Nerd Font.
			{ "nvim-tree/nvim-web-devicons", enabled = vim.g.have_nerd_font },
		},
		config = function()
			require("telescope").setup({
				extensions = {
					["ui-select"] = {
						require("telescope.themes").get_dropdown(),
					},
				},
			})

			-- Enable Telescope extensions if they are installed
			pcall(require("telescope").load_extension, "fzf")
			pcall(require("telescope").load_extension, "ui-select")

			-- See `:help telescope.builtin`
			local builtin = require("telescope.builtin")
			vim.keymap.set("n", "<leader>ff", builtin.find_files, { desc = "[F]ind [F]iles" })
			vim.keymap.set("n", "<leader>fg", builtin.live_grep, { desc = "[F]ind by [G]rep" })
			vim.keymap.set("n", "<leader>fs", builtin.grep_string, { desc = "[F]ind by [S]tring Under Cursor" })
			vim.keymap.set("n", "<leader>fp", builtin.registers, { desc = "[F]ind registers to [P]aste" })
			vim.keymap.set("n", "<leader>fr", builtin.resume, { desc = "[F]ind [R]esume" })
			vim.keymap.set("n", "<leader>f.", builtin.oldfiles, { desc = '[F]ind Recent Files ("." for repeat)' })
			vim.keymap.set("n", "<leader>fh", builtin.help_tags, { desc = "[F]ind [H]elp" })
			vim.keymap.set("n", "<leader>fb", builtin.buffers, { desc = "[F]ind existing [B]uffers" })
		end,
	},
	{
		"nvim-treesitter/nvim-treesitter",
		config = function()
			require("nvim-treesitter.configs").setup({
				ensure_installed = { "c", "lua", "vimdoc", "query", "vue", "typescript" },
				auto_install = true,
				highlight = {
					enable = true,
				},
				modules = {},
				sync_install = false,
				ignore_install = {},
				incremental_selection = {
					enable = true,
					keymaps = {
						-- start selection first then use increase or decrease to select more within
						-- i.e. <leader>ss <leader<sc> selects the scope of the function
						init_selection = "<leader>ss",
						node_incremental = "<leader>si",
						scope_incremental = "<leader>sc",
						node_decremental = "<leader>sd",
					},
				},
			})
		end,
	},
	{
		"echasnovski/mini.nvim",
		version = "*",
		config = function()
			-- require("mini.animate").setup()
			-- Better Around/Inside textobjects
			--
			-- Examples:
			--  - va)  - [V]isually select [A]round [)]paren
			--  - yinq - [Y]ank [I]nside [N]ext [Q]uote
			--  - ci'  - [C]hange [I]nside [']quote
			require("mini.ai").setup({ n_lines = 500 })
			-- Main textobject prefixes
			--   around = 'a',
			--   inside = 'i',
			--
			--   -- Next/last variants
			--   around_next = 'an',
			--   inside_next = 'in',
			--   around_last = 'al',
			--   inside_last = 'il',
			--
			--   -- Move cursor to corresponding edge of `a` textobject
			--   goto_left = 'g[',
			--   goto_right = 'g]',

			require("mini.surround").setup({
				-- Add/delete/replace surroundings (brackets, quotes, etc.)
				--
				-- - ysaiw) - [YS]urround [A]dd [I]nner [W]ord [)]Paren
				-- - ysd'   - [S]urround [D]elete [']quotes
				-- - sr)'  - [S]urround [R]eplace [)] [']
				mappings = {
					add = "ys", -- Add surrounding in Normal and Visual modes
					delete = "yd", -- Delete surrounding
					find = "yf", -- Find surrounding (to the right)
					find_left = "yF", -- Find surrounding (to the left)
					highlight = "'yh'", -- Highlight surrounding
					replace = "yr", -- Replace surrounding
					update_n_lines = "yn", -- Update `n_lines`

					suffix_last = "l", -- Suffix to search with "prev" method
					suffix_next = "n", -- Suffix to search with "next" method
				},
			})
			--    add = 'sa', -- Add surrounding in Normal and Visual modes
			--    delete = 'sd', -- Delete surrounding
			--    find = 'sf', -- Find surrounding (to the right)
			--    find_left = 'sF', -- Find surrounding (to the left)
			--    highlight = 'sh', -- Highlight surrounding
			--    replace = 'sr', -- Replace surrounding
			--    update_n_lines = 'sn', -- Update `n_lines`
			--    suffix_last = 'l', -- Suffix to search with "prev" method
			--    suffix_next = 'n', -- Suffix to search with "next" method

			local statusline = require("mini.statusline")

			statusline.setup({ use_icons = vim.g.have_nerd_font })

			---@diagnostic disable-next-line: duplicate-set-field
			statusline.section_location = function()
				return "%2l:%-2v"
			end

			require("mini.files").setup({
				content = {
					filter = nil,
					prefix = nil,
					sort = nil,
				},

				-- Module mappings created only inside explorer.
				-- Use `''` (empty string) to not create one.
				mappings = {
					close = "q",
					go_in = "l",
					go_in_plus = "<cr>",
					go_out = "h",
					go_out_plus = "H",
					mark_goto = "'",
					mark_set = "m",
					reset = "<BS>",
					reveal_cwd = "@",
					show_help = "g?",
					synchronize = "=",
					trim_left = "<",
					trim_right = ">",
				},
				options = {
					permanent_delete = true,
					use_as_default_explorer = true,
				},
				windows = {
					max_number = math.huge,
					preview = true,
					width_focus = 50,
					width_nofocus = 15,
					width_preview = 25,
				},
			})

			-- use ` to open MiniFiles
			vim.keymap.set("n", "`", function()
				local buf_name = vim.api.nvim_buf_get_name(0)
				if buf_name == "" or not vim.uv.fs_stat(buf_name) then
					-- If buffer is not a file, open mini.files with the current working directory
					MiniFiles.open(vim.fn.getcwd(), false)
				else
					local res = vim.uv.fs_stat(buf_name)
					MiniFiles.open(buf_name, false)
				end
			end, { desc = "Open File Tree" })

			local map_split = function(buf_id, keymap, direction, close)
				local open_split_and_load_file_into = function()
					-- Make new window
					local cur_target = MiniFiles.get_explorer_state().target_window
					local new_target = vim.api.nvim_win_call(cur_target, function()
						vim.cmd(direction .. " split")
						return vim.api.nvim_get_current_win()
					end)

					-- Set new window as target
					MiniFiles.set_target_window(new_target)
					MiniFiles.go_in({ close_on_file = close })
				end

				-- Adding `desc` will result into `show_help` entries
				local desc = "Split " .. direction

				vim.keymap.set("n", keymap, open_split_and_load_file_into, { buffer = buf_id, desc = desc })
			end

			local open_tabnew_and_load_file_into_it = function()
				-- Make new window
				local cur_target = MiniFiles.get_explorer_state().target_window
				local new_target = vim.api.nvim_win_call(cur_target, function()
					vim.cmd("tabnew")
					return vim.api.nvim_get_current_win()
				end)

				-- Set new window as target
				MiniFiles.set_target_window(new_target)
				MiniFiles.go_in({ close_on_file = true })
			end

			vim.api.nvim_create_autocmd("User", {
				pattern = "MiniFilesBufferCreate",
				callback = function(args)
					local buf_id = args.data.buf_id
					map_split(buf_id, "<leader>s", "belowright horizontal", true)
					map_split(buf_id, "<C-s>", "belowright horizontal", false)

					map_split(buf_id, "<leader>v", "belowright vertical", true)
					map_split(buf_id, "<C-v>", "belowright vertical", false)

					-- both open and close file explorer on command
					vim.keymap.set(
						"n",
						"<leader>t",
						open_tabnew_and_load_file_into_it,
						{ buffer = buf_id, desc = "Open file in new tab" }
					)
					vim.keymap.set(
						"n",
						"<C-t>",
						open_tabnew_and_load_file_into_it,
						{ buffer = buf_id, desc = "Open file in new tab" }
					)
				end,
			})
		end,
	},
	{
		"CopilotC-Nvim/CopilotChat.nvim",
		dependencies = {
			{ "github/copilot.vim" }, -- or zbirenbaum/copilot.lua
			{ "nvim-lua/plenary.nvim", branch = "master" }, -- for curl, log and async functions
		},
		build = "make tiktoken", -- Only on MacOS or Linux
		opts = {
			-- See Configuration section for options
		},
	},
	{
		"claydugo/browsher.nvim",
		event = "VeryLazy",
		config = function()
			require("browsher").setup({
				default_pin = "branch",
				open_cmd = "open",
				allow_line_numbers_with_uncommitted_changes = true,
			})
		end,
	},
	{
		"https://github.com/aaronik/treewalker.nvim",
		opts = {
			highlight = true,
		},
	},
	{ -- annotations like tsdoc
		"danymat/neogen",
		config = function()
			local opts = { noremap = true, silent = true }
			vim.api.nvim_set_keymap("n", "<leader>nf", ":lua require('neogen').generate()<CR>", opts)
			require("neogen").setup({
				enabled = true, --if you want to disable Neogen
				input_after_comment = true, -- (default: true) automatic jump (with insert mode) on inserted annotation
			})
		end,
		version = "*",
	},
	"sindrets/diffview.nvim",
	{
		"mistweaverco/kulala.nvim",
		keys = {
			{ "<leader>Rs", desc = "Send request" },
			{ "<leader>Ra", desc = "Send all requests" },
			{ "<leader>Rb", desc = "Open scratchpad" },
		},
		ft = { "http", "rest" },
		opts = {
			global_keymaps = true,
			global_keymaps_prefix = "<leader>R",
			kulala_keymaps_prefix = "",
			halt_on_error = false, -- if true, will stop processing requests on error
		},
	},
	lsp_plugin_table,
})
