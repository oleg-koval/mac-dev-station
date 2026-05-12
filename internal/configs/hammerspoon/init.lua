-- Hammerspoon configuration
-- https://www.hammerspoon.org/

-- Load display watcher
dofile(os.getenv("HOME") .. "/.hammerspoon/display-watcher.lua")

-- Auto-reload config on changes
hs.pathwatcher.new(os.getenv("HOME") .. "/.hammerspoon/", function(files)
    hs.reload()
end):start()

-- Notification when config reloads
function reloadConfig(files)
    dofile(os.getenv("HOME") .. "/.hammerspoon/init.lua")
    hs.notify.new({title="Hammerspoon", informativeText="Config loaded"}):send()
end

-- Logger
logger = hs.logger.new('hammerspoon', 'debug')

-- Alert styles
function alert(msg)
    hs.alert.show(msg, 2)
end

-- Initial log
logger:i("Hammerspoon config loaded")
alert("Hammerspoon ready")
