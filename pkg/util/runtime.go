package util

// DryRun, when true, prevents any mutating operations (downloads, deletions)
// and instead logs what would happen.
var DryRun bool

// ShowProgressBars controls whether progress bars are displayed during downloads.
// Defaults to true; can be disabled via CLI flag.
var ShowProgressBars = true
