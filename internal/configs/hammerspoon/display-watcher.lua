-- Display dock/undock auto-detection
-- Watches for external displays being connected/disconnected

local logger = hs.logger.new('display-watcher', 'debug')

-- Callback when screens change
function onScreenChange()
    local screens = hs.screen.allScreens()
    local screenCount = #screens

    logger:i("Screen configuration changed. Screen count: " .. screenCount)

    -- Log current screen info
    for i, screen in ipairs(screens) do
        local res = screen:fullFrame()
        logger:i("Screen " .. i .. ": " .. screen:name() .. " - " ..
                 tostring(res.w) .. "x" .. tostring(res.h))
    end

    -- Trigger notification
    hs.notify.new({
        title = "Display Changed",
        informativeText = "Screen count: " .. screenCount
    }):send()
end

-- Create screen watcher
screenWatcher = hs.screen.watcher.new(onScreenChange):start()

logger:i("Display watcher initialized")
