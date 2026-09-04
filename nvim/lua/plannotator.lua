local tmux = require("plannotator.tmux")

local M = {}

local CHROME_APP_BROWSER = vim.fn.expand("~/.scratch/scripts/plannotator-chrome-app")

local config = {
	submit = false,
	focus = true,
}

local function resolve_browser()
	if vim.fn.executable(CHROME_APP_BROWSER) == 1 then
		return CHROME_APP_BROWSER
	end
	return vim.env.PLANNOTATOR_BROWSER
end

local function current_file()
	local path = vim.api.nvim_buf_get_name(0)
	if path == "" then
		return nil
	end
	if vim.bo.modified then
		vim.cmd("write")
	end
	return path
end

local function show_in_scratch(payload)
	vim.cmd("botright split")
	local buffer = vim.api.nvim_create_buf(false, true)
	vim.api.nvim_win_set_buf(0, buffer)
	vim.api.nvim_buf_set_lines(buffer, 0, -1, false, vim.split(payload, "\n"))
	vim.bo[buffer].buftype = "nofile"
	vim.bo[buffer].bufhidden = "wipe"
	vim.bo[buffer].filetype = "markdown"
	vim.bo[buffer].modifiable = false
	vim.keymap.set("n", "q", "<cmd>close<cr>", { buffer = buffer, nowait = true })
end

local function build_payload(feedback, path)
	local relative = vim.fn.fnamemodify(path, ":.")
	return string.format("Plannotator feedback on `%s`:\n\n%s", relative, feedback)
end

local function deliver(feedback, path, pane)
	local payload = build_payload(feedback, path)
	vim.fn.setreg("+", payload)
	vim.fn.setreg('"', payload)

	if pane and tmux.is_alive(pane) then
		tmux.paste(pane, payload, config)
		vim.notify("Plannotator: feedback sent to " .. pane.target)
		return
	end

	vim.notify("Plannotator: no opencode pane available, feedback copied to clipboard")
	show_in_scratch(payload)
end

local function on_exit(path, pane)
	return vim.schedule_wrap(function(result)
		local feedback = vim.trim(result.stdout or "")
		if feedback == "" then
			vim.notify("Plannotator: no feedback returned")
			return
		end
		deliver(feedback, path, pane)
	end)
end

local function launch(target, pane)
	vim.notify("Plannotator: opening " .. vim.fn.fnamemodify(target, ":t") .. "…")
	vim.system({ "plannotator", "annotate", target, "--gate" }, {
		env = { PLANNOTATOR_BROWSER = resolve_browser() },
		text = true,
	}, on_exit(target, pane))
end

--- Opens `path` (defaults to the current buffer) in the Plannotator annotation UI.
--- The destination opencode pane is picked up front, before the browser opens.
function M.annotate(path)
	if vim.fn.executable("plannotator") == 0 then
		vim.notify("Plannotator: `plannotator` is not on PATH", vim.log.levels.ERROR)
		return
	end

	local target = path or current_file()
	if not target then
		vim.notify("Plannotator: buffer has no file on disk", vim.log.levels.ERROR)
		return
	end

	tmux.pick(target, function(pane)
		launch(target, pane)
	end)
end

function M.setup(options)
	config = vim.tbl_extend("force", config, options or {})
end

vim.api.nvim_create_user_command("Plannotator", function(args)
	M.annotate(args.args ~= "" and args.args or nil)
end, { nargs = "?", complete = "file", desc = "Annotate a file, folder or URL in Plannotator" })

return M
