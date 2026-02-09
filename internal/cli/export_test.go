package cli

// Export unexported functions for testing
var ValidateDictFile = validateDictFile
var CopyFile = copyFile
var DetectLanguagePair = detectLanguagePair
var HighlightWord = highlightWord
var FormatSubjects = formatSubjects
var GroupByWordType = groupByWordType
var WordTypeLabel = wordTypeLabel
var WordTypeColor = wordTypeColor
var GetTermWidth = &getTermWidth
var ClampColumnWidths = clampColumnWidths
var TruncateCell = truncateCell
