package southpark

type excludeRule struct {
	Language          Language
	Season            int
	Episode           int
	SkipVideoSegments []int
}

// HACK: Some episodes' segments are broken in a way that I can't
// check before muxing each segment into the video. Therefore I'm
// just manually adding them here.
var excludeRules = []excludeRule{
	{LanguageSpanish, 4, 4, []int{83, 118, 167, 239}},
}
