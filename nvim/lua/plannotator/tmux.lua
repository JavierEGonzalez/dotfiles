local M = {}

local PANE_FORMAT = table.concat({
	"#{pane_id}",
	"#{session_name}:#{window_index}.#{pane_index}",
	"#{pane_current_command}",
	"#{pane_current_path}",
}, "\t")

local AGENT_COMMAND_PATTERN = "^opencode"
local TMUX_BUFFER_NAME = "plannotator"

local function inside_tmux()
	return vim.env.TMUX ~= nil and vim.env.TMUX ~= ""
end

local function parse_pane(line)
	local id, target, command, path = line:match("^(.-)\t(.-)\t(.-)\t(.*)$")
	if not id then
		return nil
	end
	return { id = id, target = target, command = command, path = path }
end

--- Lists every tmux pane currently running an opencode agent.
function M.list_agent_panes()
	if vim.fn.executable("tmux") == 0 then
		return {}
	end

	local result = vim.system({ "tmux", "list-panes", "-a", "-F", PANE_FORMAT }, { text = true }):wait()
	if result.code ~= 0 then
		return {}
	end

	local panes = {}
	for line in (result.stdout or ""):gmatch("[^\n]+") do
		local pane = parse_pane(line)
		if pane and pane.command:match(AGENT_COMMAND_PATTERN) then
			table.insert(panes, pane)
		end
	end
	return panes
end

local function is_ancestor_of(directory, file_path)
	return directory ~= "" and file_path:sub(1, #directory + 1) == directory .. "/"
end

--- Sorts panes whose working directory contains `file_path` to the front.
local function prefer_panes_near(panes, file_path)
	table.sort(panes, function(left, right)
		local left_matches = is_ancestor_of(left.path, file_path)
		local right_matches = is_ancestor_of(right.path, file_path)
		if left_matches ~= right_matches then
			return left_matches
		end
		return #left.path > #right.path
	end)
	return panes
end

local function describe(pane)
	return string.format("%s — %s", pane.target, vim.fn.fnamemodify(pane.path, ":t"))
end

local function choose_pane(file_path, on_choice)
	local panes = prefer_panes_near(M.list_agent_panes(), file_path)
	if #panes == 0 then
		return on_choice(nil)
	end
	if #panes == 1 then
		return on_choice(panes[1])
	end
	vim.ui.select(panes, { prompt = "Plannotator → opencode pane", format_item = describe }, on_choice)
end

--- Asks which opencode pane should receive the feedback.
--- Calls `on_choice(pane)` with the chosen pane, or nil when none is available.
function M.pick(file_path, on_choice)
	choose_pane(file_path, on_choice)
end

--- Reports whether a previously picked pane still exists.
function M.is_alive(pane)
	local result = vim.system({ "tmux", "list-panes", "-a", "-F", "#{pane_id}" }, { text = true }):wait()
	for id in (result.stdout or ""):gmatch("[^\n]+") do
		if id == pane.id then
			return true
		end
	end
	return false
end

--- Pastes `payload` into `pane` without submitting it, unless `options.submit`.
function M.paste(pane, payload, options)
	vim.system({ "tmux", "load-buffer", "-b", TMUX_BUFFER_NAME, "-" }, { stdin = payload }):wait()
	vim.system({ "tmux", "paste-buffer", "-b", TMUX_BUFFER_NAME, "-d", "-p", "-t", pane.id }):wait()

	if options.submit then
		vim.system({ "tmux", "send-keys", "-t", pane.id, "Enter" }):wait()
	end
	if options.focus and inside_tmux() then
		vim.system({ "tmux", "switch-client", "-t", pane.id }):wait()
	end
end

return M
