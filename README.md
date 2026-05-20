# FDicM

> Blazing fast responsive split pane terminal dictionary app engine built with Go and Bubble Tea

---

## Architecture

FDicM delivers a localized cross platform dictionary application interface designed directly for keyboard driven terminal environments. The framework operates with zero external dynamic runtime link dependencies, utilizing a dual panel view system to maximize text layout visibility and parsing efficiency while reading raw lexeme metadata definitions.

---

## Performance

**Responsive Layout Engine**
Terminal window resizing operations are handled dynamically to adjust viewport heights cleanly without text clipping artifacts.

**Dual Source Data Resolution**
Primary lookups execute parallel network queries targeting cloud vocabulary endpoints. On network dropout or timeout, the engine automatically falls back to a local compressed dictionary database.

**Low Memory Footprint**
Static memory allocation strategies keep runtime usage exceptionally low during intensive vocabulary lookups.

---

## Keybindings

| Key   | Action                                                                 |
|-------|------------------------------------------------------------------------|
| TAB   | Toggle focus between the query prompt and the definition viewport      |
| p     | Stream remote audio pronunciation or invoke native text to speech      |
| v     | Toggle syntax colored JSON visualization in the secondary panel        |
| c     | Copy the clean definition string to the system clipboard               |
| q     | Close the active viewport or terminate the application                 |
| ESC   | Exit secondary panels or interrupt foreground rendering loops          |
| j     | Scroll down within the focused definition viewport                     |
| k     | Scroll up within the focused definition viewport                       |

---

## Usage

**Launch**

    fdicm

**Filter by word class**

    :noun interface
    :verb program

**Check version**

    fdicm --version

---

## Installation

**macOS via Homebrew**

    brew tap jagath-sajjan/tap
    brew install fdicm

