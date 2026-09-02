## 2026.09.02
- Version Bump for Gotify 3.1.0

## 2026.08.15
- Update build version to Gotify 3,0.0
- Changed method of aquiring client token from using LocalStorage to use the value found in the Cookie.
- API usage now uses cookie rather than header.
- Removed Wrapper HTML due to no longer needing to pull value from LocalStorage.
- Clients are now named with "unique" names for selection purposes.
- Clients are no longer deleted. Instead they are renamed and have an experation applied. Clients in the future may have a inactivity timer set rather than being removed in this way.

## 2026.05.04
- Version Bump for Gotify 2.9.1

## 2026.02.14
- Version Bump for Gotify 2.9.0

## 2026.01.05
- Version Bump for Gotify 2.8.0

## 2025.09.21
- Version Bump for Gotify 2.7.3

## 2025.09.13
- Version Bump for Gotify 2.7.2

## 2025.08.10
First Release
### New Features
- Ability to add feeds to a watch list
- Ability to remove feeds from a watch list
- Detection and publication of new feed items automatically once every 5 minutes if available.
