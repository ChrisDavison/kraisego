# Kraise

This is a run-or-raise tool for use in KDE wayland, as a wrapper around `kdotool`.

## Usage

```bash
kraisego -c firefox -e YouTube   # find a firefox window EXCLUDING those with title matching YouTube

kraisego -c firefox -t 'YouTube|Netflix'   # find firefox window with YouTube or netflix in title

kraisego -c anki -l anki    # find any anki window, otherwise launch anki
```

## Installation

```bash
go install github.com/chrisdavison/kraisego
```
