# CEKlopfenstein/simple-feeds "Simple Feeds"

[![Release Pipeline](https://github.com/CEKlopfenstein/simple-feeds/actions/workflows/release.yml/badge.svg?branch=master)](https://github.com/CEKlopfenstein/simple-feeds/actions/workflows/release.yml)

A simple Gotify Plugin that periodically queries different feeds and publishes new feed items to the attached Gotify Server.

## Features
- Graphical User Interface
   - Allows the management of feeds (Add/Delete)
- Support for multiple feed types
   - Support of multiple feed types is achieved through [gofeed](https://github.com/mmcdole/gofeed) library.
- Able to determine whether feed items are "new" through different means.
   - If the feed in question does not provide a timestamp for when posts are made, this plugin will fall back on attempting to use the URL of the feed item in question to determine "newness".

## Motivation
I previously was using a Chrome plugin called RSS Feed Reader to watch RSS feeds. But I tend to miss the emails it sends out. I wanted to be able to place the updates of the feeds into a Discord Server. Which can be done using this plugin and my [Gotify Relay](https://github.com/CEKlopfenstein/gotify-repeater) plugin.

## Currently Planned Features
- Add the ability to have separate Gotify "Apps" for separate feeds.
   - Currently, all feeds go into a single app for the plugin itself.

## [Changelog](/CHANGELOG.md)

## Installation
1. Download the [latest "stable" version](https://github.com/CEKlopfenstein/simple-feeds/releases/latest) for your desired deployment.
   > Note: As of writing I only actively test the AMD64 build. ARM64 and ARM-7 should work. But are not garenteed.
2. Place the downloaded *.so file within your Gotify instance's plugins folder.
   > Default configuration for Docker will find it at `/app/data/plugins/` within the container.
   
   > Limited to Linux and MacOS. [Gotify documentation mentioning the limitation](https://gotify.net/docs/plugin)
3. Start/Restart your Gotify instance. (Required for the plugin to be loaded.)
4. Login to your Gotify instance and navigate to plugs.
   > ![](/images/plugins.png)
5. Enable the Simple Feeds plugin and navigate to the plugin info page.
   > ![](/images/info.png)
6. Click either the Route Prefix or the Config Page link.
7. Click either Use Current Client Token or Create Custom Token.
   > ![](/images/plugin_config.png)
   
   > Tokens are currently not used and can be safely ignored for now. But may be required for future features.
8. Add feeds as desired using the UI.
   > Feeds can also be deleted from this view.

## Building From Source
For now please refer to [OG_README.md](OG_README.md) for documentation on how to build. Cloning the repository and running `make build` "should" work. But is not guarenteed. As it was modified to function on my machine due to some strange issues. And due to `docker` not being configured to be accessible without `sudo` on my machine.

Within the Makefile is also an option to run `make run` which will build the AMD64 version of the plugin and deploy it onto a local instance of Gotify in a Docker container using [gotify/server](https://github.com/gotify/server).

## Other Notes
The original [README](/OG_README.md) from [gotify/plugin-templates](https://github.com/gotify/plugin-template) is avaliable as `OG_README.md`.