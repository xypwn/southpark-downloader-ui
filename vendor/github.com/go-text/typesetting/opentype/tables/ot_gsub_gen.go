// SPDX-License-Identifier: Unlicense OR BSD-3-Clause

package tables

import (
	"encoding/binary"
	"fmt"
)

// Code generated by binarygen from ot_gsub_src.go. DO NOT EDIT

func ParseAlternateSet(src []byte) (AlternateSet, int, error) {
	var item AlternateSet
	n := 0
	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading AlternateSet: "+"EOF: expected length: 2, got %d", L)
	}
	arrayLengthAlternateGlyphIDs := int(binary.BigEndian.Uint16(src[0:]))
	n += 2

	{

		if L := len(src); L < 2+arrayLengthAlternateGlyphIDs*2 {
			return item, 0, fmt.Errorf("reading AlternateSet: "+"EOF: expected length: %d, got %d", 2+arrayLengthAlternateGlyphIDs*2, L)
		}

		item.AlternateGlyphIDs = make([]uint16, arrayLengthAlternateGlyphIDs) // allocation guarded by the previous check
		for i := range item.AlternateGlyphIDs {
			item.AlternateGlyphIDs[i] = binary.BigEndian.Uint16(src[2+i*2:])
		}
		n += arrayLengthAlternateGlyphIDs * 2
	}
	return item, n, nil
}

func ParseAlternateSubs(src []byte) (AlternateSubs, int, error) {
	var item AlternateSubs
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading AlternateSubs: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.substFormat = binary.BigEndian.Uint16(src[0:])
	offsetCoverage := int(binary.BigEndian.Uint16(src[2:]))
	arrayLengthAlternateSets := int(binary.BigEndian.Uint16(src[4:]))
	n += 6

	{

		if offsetCoverage != 0 { // ignore null offset
			if L := len(src); L < offsetCoverage {
				return item, 0, fmt.Errorf("reading AlternateSubs: "+"EOF: expected length: %d, got %d", offsetCoverage, L)
			}

			var (
				err  error
				read int
			)
			item.Coverage, read, err = ParseCoverage(src[offsetCoverage:])
			if err != nil {
				return item, 0, fmt.Errorf("reading AlternateSubs: %s", err)
			}
			offsetCoverage += read
		}
	}
	{

		if L := len(src); L < 6+arrayLengthAlternateSets*2 {
			return item, 0, fmt.Errorf("reading AlternateSubs: "+"EOF: expected length: %d, got %d", 6+arrayLengthAlternateSets*2, L)
		}

		item.AlternateSets = make([]AlternateSet, arrayLengthAlternateSets) // allocation guarded by the previous check
		for i := range item.AlternateSets {
			offset := int(binary.BigEndian.Uint16(src[6+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading AlternateSubs: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.AlternateSets[i], _, err = ParseAlternateSet(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading AlternateSubs: %s", err)
			}
		}
		n += arrayLengthAlternateSets * 2
	}
	return item, n, nil
}

func ParseChainedContextualSubs(src []byte) (ChainedContextualSubs, int, error) {
	var item ChainedContextualSubs
	n := 0
	{
		var (
			err  error
			read int
		)
		item.Data, read, err = ParseChainedContextualSubsITF(src[0:])
		if err != nil {
			return item, 0, fmt.Errorf("reading ChainedContextualSubs: %s", err)
		}
		n += read
	}
	return item, n, nil
}

func ParseChainedContextualSubs1(src []byte) (ChainedContextualSubs1, int, error) {
	var item ChainedContextualSubs1
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading ChainedContextualSubs1: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.format = binary.BigEndian.Uint16(src[0:])
	offsetCoverage := int(binary.BigEndian.Uint16(src[2:]))
	arrayLengthChainedSeqRuleSet := int(binary.BigEndian.Uint16(src[4:]))
	n += 6

	{

		if offsetCoverage != 0 { // ignore null offset
			if L := len(src); L < offsetCoverage {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs1: "+"EOF: expected length: %d, got %d", offsetCoverage, L)
			}

			var (
				err  error
				read int
			)
			item.coverage, read, err = ParseCoverage(src[offsetCoverage:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs1: %s", err)
			}
			offsetCoverage += read
		}
	}
	{

		if L := len(src); L < 6+arrayLengthChainedSeqRuleSet*2 {
			return item, 0, fmt.Errorf("reading ChainedContextualSubs1: "+"EOF: expected length: %d, got %d", 6+arrayLengthChainedSeqRuleSet*2, L)
		}

		item.ChainedSeqRuleSet = make([]ChainedSequenceRuleSet, arrayLengthChainedSeqRuleSet) // allocation guarded by the previous check
		for i := range item.ChainedSeqRuleSet {
			offset := int(binary.BigEndian.Uint16(src[6+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs1: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.ChainedSeqRuleSet[i], _, err = ParseChainedSequenceRuleSet(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs1: %s", err)
			}
		}
		n += arrayLengthChainedSeqRuleSet * 2
	}
	return item, n, nil
}

func ParseChainedContextualSubs2(src []byte) (ChainedContextualSubs2, int, error) {
	var item ChainedContextualSubs2
	n := 0
	if L := len(src); L < 12 {
		return item, 0, fmt.Errorf("reading ChainedContextualSubs2: "+"EOF: expected length: 12, got %d", L)
	}
	_ = src[11] // early bound checking
	item.format = binary.BigEndian.Uint16(src[0:])
	offsetCoverage := int(binary.BigEndian.Uint16(src[2:]))
	offsetBacktrackClassDef := int(binary.BigEndian.Uint16(src[4:]))
	offsetInputClassDef := int(binary.BigEndian.Uint16(src[6:]))
	offsetLookaheadClassDef := int(binary.BigEndian.Uint16(src[8:]))
	arrayLengthChainedClassSeqRuleSet := int(binary.BigEndian.Uint16(src[10:]))
	n += 12

	{

		if offsetCoverage != 0 { // ignore null offset
			if L := len(src); L < offsetCoverage {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs2: "+"EOF: expected length: %d, got %d", offsetCoverage, L)
			}

			var (
				err  error
				read int
			)
			item.coverage, read, err = ParseCoverage(src[offsetCoverage:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs2: %s", err)
			}
			offsetCoverage += read
		}
	}
	{

		if offsetBacktrackClassDef != 0 { // ignore null offset
			if L := len(src); L < offsetBacktrackClassDef {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs2: "+"EOF: expected length: %d, got %d", offsetBacktrackClassDef, L)
			}

			var (
				err  error
				read int
			)
			item.BacktrackClassDef, read, err = ParseClassDef(src[offsetBacktrackClassDef:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs2: %s", err)
			}
			offsetBacktrackClassDef += read
		}
	}
	{

		if offsetInputClassDef != 0 { // ignore null offset
			if L := len(src); L < offsetInputClassDef {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs2: "+"EOF: expected length: %d, got %d", offsetInputClassDef, L)
			}

			var (
				err  error
				read int
			)
			item.InputClassDef, read, err = ParseClassDef(src[offsetInputClassDef:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs2: %s", err)
			}
			offsetInputClassDef += read
		}
	}
	{

		if offsetLookaheadClassDef != 0 { // ignore null offset
			if L := len(src); L < offsetLookaheadClassDef {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs2: "+"EOF: expected length: %d, got %d", offsetLookaheadClassDef, L)
			}

			var (
				err  error
				read int
			)
			item.LookaheadClassDef, read, err = ParseClassDef(src[offsetLookaheadClassDef:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs2: %s", err)
			}
			offsetLookaheadClassDef += read
		}
	}
	{

		if L := len(src); L < 12+arrayLengthChainedClassSeqRuleSet*2 {
			return item, 0, fmt.Errorf("reading ChainedContextualSubs2: "+"EOF: expected length: %d, got %d", 12+arrayLengthChainedClassSeqRuleSet*2, L)
		}

		item.ChainedClassSeqRuleSet = make([]ChainedSequenceRuleSet, arrayLengthChainedClassSeqRuleSet) // allocation guarded by the previous check
		for i := range item.ChainedClassSeqRuleSet {
			offset := int(binary.BigEndian.Uint16(src[12+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs2: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.ChainedClassSeqRuleSet[i], _, err = ParseChainedSequenceRuleSet(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs2: %s", err)
			}
		}
		n += arrayLengthChainedClassSeqRuleSet * 2
	}
	return item, n, nil
}

func ParseChainedContextualSubs3(src []byte) (ChainedContextualSubs3, int, error) {
	var item ChainedContextualSubs3
	n := 0
	if L := len(src); L < 4 {
		return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: 4, got %d", L)
	}
	_ = src[3] // early bound checking
	item.format = binary.BigEndian.Uint16(src[0:])
	arrayLengthBacktrackCoverages := int(binary.BigEndian.Uint16(src[2:]))
	n += 4

	{

		if L := len(src); L < 4+arrayLengthBacktrackCoverages*2 {
			return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: %d, got %d", 4+arrayLengthBacktrackCoverages*2, L)
		}

		item.BacktrackCoverages = make([]Coverage, arrayLengthBacktrackCoverages) // allocation guarded by the previous check
		for i := range item.BacktrackCoverages {
			offset := int(binary.BigEndian.Uint16(src[4+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.BacktrackCoverages[i], _, err = ParseCoverage(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs3: %s", err)
			}
		}
		n += arrayLengthBacktrackCoverages * 2
	}
	if L := len(src); L < n+2 {
		return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: n + 2, got %d", L)
	}
	arrayLengthInputCoverages := int(binary.BigEndian.Uint16(src[n:]))
	n += 2

	{

		if L := len(src); L < n+arrayLengthInputCoverages*2 {
			return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: %d, got %d", n+arrayLengthInputCoverages*2, L)
		}

		item.InputCoverages = make([]Coverage, arrayLengthInputCoverages) // allocation guarded by the previous check
		for i := range item.InputCoverages {
			offset := int(binary.BigEndian.Uint16(src[n+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.InputCoverages[i], _, err = ParseCoverage(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs3: %s", err)
			}
		}
		n += arrayLengthInputCoverages * 2
	}
	if L := len(src); L < n+2 {
		return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: n + 2, got %d", L)
	}
	arrayLengthLookaheadCoverages := int(binary.BigEndian.Uint16(src[n:]))
	n += 2

	{

		if L := len(src); L < n+arrayLengthLookaheadCoverages*2 {
			return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: %d, got %d", n+arrayLengthLookaheadCoverages*2, L)
		}

		item.LookaheadCoverages = make([]Coverage, arrayLengthLookaheadCoverages) // allocation guarded by the previous check
		for i := range item.LookaheadCoverages {
			offset := int(binary.BigEndian.Uint16(src[n+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.LookaheadCoverages[i], _, err = ParseCoverage(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ChainedContextualSubs3: %s", err)
			}
		}
		n += arrayLengthLookaheadCoverages * 2
	}
	if L := len(src); L < n+2 {
		return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: n + 2, got %d", L)
	}
	arrayLengthSeqLookupRecords := int(binary.BigEndian.Uint16(src[n:]))
	n += 2

	{

		if L := len(src); L < n+arrayLengthSeqLookupRecords*4 {
			return item, 0, fmt.Errorf("reading ChainedContextualSubs3: "+"EOF: expected length: %d, got %d", n+arrayLengthSeqLookupRecords*4, L)
		}

		item.SeqLookupRecords = make([]SequenceLookupRecord, arrayLengthSeqLookupRecords) // allocation guarded by the previous check
		for i := range item.SeqLookupRecords {
			item.SeqLookupRecords[i].mustParse(src[n+i*4:])
		}
		n += arrayLengthSeqLookupRecords * 4
	}
	return item, n, nil
}

func ParseChainedContextualSubsITF(src []byte) (ChainedContextualSubsITF, int, error) {
	var item ChainedContextualSubsITF

	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading ChainedContextualSubsITF: "+"EOF: expected length: 2, got %d", L)
	}
	format := uint16(binary.BigEndian.Uint16(src[0:]))
	var (
		read int
		err  error
	)
	switch format {
	case 1:
		item, read, err = ParseChainedContextualSubs1(src[0:])
	case 2:
		item, read, err = ParseChainedContextualSubs2(src[0:])
	case 3:
		item, read, err = ParseChainedContextualSubs3(src[0:])
	default:
		err = fmt.Errorf("unsupported ChainedContextualSubsITF format %d", format)
	}
	if err != nil {
		return item, 0, fmt.Errorf("reading ChainedContextualSubsITF: %s", err)
	}

	return item, read, nil
}

func ParseContextualSubs(src []byte) (ContextualSubs, int, error) {
	var item ContextualSubs
	n := 0
	{
		var (
			err  error
			read int
		)
		item.Data, read, err = ParseContextualSubsITF(src[0:])
		if err != nil {
			return item, 0, fmt.Errorf("reading ContextualSubs: %s", err)
		}
		n += read
	}
	return item, n, nil
}

func ParseContextualSubs1(src []byte) (ContextualSubs1, int, error) {
	var item ContextualSubs1
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading ContextualSubs1: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.format = binary.BigEndian.Uint16(src[0:])
	offsetCoverage := int(binary.BigEndian.Uint16(src[2:]))
	arrayLengthSeqRuleSet := int(binary.BigEndian.Uint16(src[4:]))
	n += 6

	{

		if offsetCoverage != 0 { // ignore null offset
			if L := len(src); L < offsetCoverage {
				return item, 0, fmt.Errorf("reading ContextualSubs1: "+"EOF: expected length: %d, got %d", offsetCoverage, L)
			}

			var (
				err  error
				read int
			)
			item.coverage, read, err = ParseCoverage(src[offsetCoverage:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ContextualSubs1: %s", err)
			}
			offsetCoverage += read
		}
	}
	{

		if L := len(src); L < 6+arrayLengthSeqRuleSet*2 {
			return item, 0, fmt.Errorf("reading ContextualSubs1: "+"EOF: expected length: %d, got %d", 6+arrayLengthSeqRuleSet*2, L)
		}

		item.SeqRuleSet = make([]SequenceRuleSet, arrayLengthSeqRuleSet) // allocation guarded by the previous check
		for i := range item.SeqRuleSet {
			offset := int(binary.BigEndian.Uint16(src[6+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ContextualSubs1: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.SeqRuleSet[i], _, err = ParseSequenceRuleSet(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ContextualSubs1: %s", err)
			}
		}
		n += arrayLengthSeqRuleSet * 2
	}
	return item, n, nil
}

func ParseContextualSubs2(src []byte) (ContextualSubs2, int, error) {
	var item ContextualSubs2
	n := 0
	if L := len(src); L < 8 {
		return item, 0, fmt.Errorf("reading ContextualSubs2: "+"EOF: expected length: 8, got %d", L)
	}
	_ = src[7] // early bound checking
	item.format = binary.BigEndian.Uint16(src[0:])
	offsetCoverage := int(binary.BigEndian.Uint16(src[2:]))
	offsetClassDef := int(binary.BigEndian.Uint16(src[4:]))
	arrayLengthClassSeqRuleSet := int(binary.BigEndian.Uint16(src[6:]))
	n += 8

	{

		if offsetCoverage != 0 { // ignore null offset
			if L := len(src); L < offsetCoverage {
				return item, 0, fmt.Errorf("reading ContextualSubs2: "+"EOF: expected length: %d, got %d", offsetCoverage, L)
			}

			var (
				err  error
				read int
			)
			item.coverage, read, err = ParseCoverage(src[offsetCoverage:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ContextualSubs2: %s", err)
			}
			offsetCoverage += read
		}
	}
	{

		if offsetClassDef != 0 { // ignore null offset
			if L := len(src); L < offsetClassDef {
				return item, 0, fmt.Errorf("reading ContextualSubs2: "+"EOF: expected length: %d, got %d", offsetClassDef, L)
			}

			var (
				err  error
				read int
			)
			item.ClassDef, read, err = ParseClassDef(src[offsetClassDef:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ContextualSubs2: %s", err)
			}
			offsetClassDef += read
		}
	}
	{

		if L := len(src); L < 8+arrayLengthClassSeqRuleSet*2 {
			return item, 0, fmt.Errorf("reading ContextualSubs2: "+"EOF: expected length: %d, got %d", 8+arrayLengthClassSeqRuleSet*2, L)
		}

		item.ClassSeqRuleSet = make([]SequenceRuleSet, arrayLengthClassSeqRuleSet) // allocation guarded by the previous check
		for i := range item.ClassSeqRuleSet {
			offset := int(binary.BigEndian.Uint16(src[8+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ContextualSubs2: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.ClassSeqRuleSet[i], _, err = ParseSequenceRuleSet(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ContextualSubs2: %s", err)
			}
		}
		n += arrayLengthClassSeqRuleSet * 2
	}
	return item, n, nil
}

func ParseContextualSubs3(src []byte) (ContextualSubs3, int, error) {
	var item ContextualSubs3
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading ContextualSubs3: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.format = binary.BigEndian.Uint16(src[0:])
	item.glyphCount = binary.BigEndian.Uint16(src[2:])
	item.seqLookupCount = binary.BigEndian.Uint16(src[4:])
	n += 6

	{
		arrayLength := int(item.glyphCount)

		if L := len(src); L < 6+arrayLength*2 {
			return item, 0, fmt.Errorf("reading ContextualSubs3: "+"EOF: expected length: %d, got %d", 6+arrayLength*2, L)
		}

		item.Coverages = make([]Coverage, arrayLength) // allocation guarded by the previous check
		for i := range item.Coverages {
			offset := int(binary.BigEndian.Uint16(src[6+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ContextualSubs3: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.Coverages[i], _, err = ParseCoverage(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ContextualSubs3: %s", err)
			}
		}
		n += arrayLength * 2
	}
	{
		arrayLength := int(item.seqLookupCount)

		if L := len(src); L < n+arrayLength*4 {
			return item, 0, fmt.Errorf("reading ContextualSubs3: "+"EOF: expected length: %d, got %d", n+arrayLength*4, L)
		}

		item.SeqLookupRecords = make([]SequenceLookupRecord, arrayLength) // allocation guarded by the previous check
		for i := range item.SeqLookupRecords {
			item.SeqLookupRecords[i].mustParse(src[n+i*4:])
		}
		n += arrayLength * 4
	}
	return item, n, nil
}

func ParseContextualSubsITF(src []byte) (ContextualSubsITF, int, error) {
	var item ContextualSubsITF

	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading ContextualSubsITF: "+"EOF: expected length: 2, got %d", L)
	}
	format := uint16(binary.BigEndian.Uint16(src[0:]))
	var (
		read int
		err  error
	)
	switch format {
	case 1:
		item, read, err = ParseContextualSubs1(src[0:])
	case 2:
		item, read, err = ParseContextualSubs2(src[0:])
	case 3:
		item, read, err = ParseContextualSubs3(src[0:])
	default:
		err = fmt.Errorf("unsupported ContextualSubsITF format %d", format)
	}
	if err != nil {
		return item, 0, fmt.Errorf("reading ContextualSubsITF: %s", err)
	}

	return item, read, nil
}

func ParseExtensionSubs(src []byte) (ExtensionSubs, int, error) {
	var item ExtensionSubs
	n := 0
	if L := len(src); L < 8 {
		return item, 0, fmt.Errorf("reading ExtensionSubs: "+"EOF: expected length: 8, got %d", L)
	}
	_ = src[7] // early bound checking
	item.substFormat = binary.BigEndian.Uint16(src[0:])
	item.ExtensionLookupType = binary.BigEndian.Uint16(src[2:])
	item.ExtensionOffset = Offset32(binary.BigEndian.Uint32(src[4:]))
	n += 8

	{

		item.RawData = src[0:]
		n = len(src)
	}
	return item, n, nil
}

func ParseLigature(src []byte) (Ligature, int, error) {
	var item Ligature
	n := 0
	if L := len(src); L < 4 {
		return item, 0, fmt.Errorf("reading Ligature: "+"EOF: expected length: 4, got %d", L)
	}
	_ = src[3] // early bound checking
	item.LigatureGlyph = binary.BigEndian.Uint16(src[0:])
	item.componentCount = binary.BigEndian.Uint16(src[2:])
	n += 4

	{
		arrayLength := int(item.componentCount - 1)

		if L := len(src); L < 4+arrayLength*2 {
			return item, 0, fmt.Errorf("reading Ligature: "+"EOF: expected length: %d, got %d", 4+arrayLength*2, L)
		}

		item.ComponentGlyphIDs = make([]uint16, arrayLength) // allocation guarded by the previous check
		for i := range item.ComponentGlyphIDs {
			item.ComponentGlyphIDs[i] = binary.BigEndian.Uint16(src[4+i*2:])
		}
		n += arrayLength * 2
	}
	return item, n, nil
}

func ParseLigatureSet(src []byte) (LigatureSet, int, error) {
	var item LigatureSet
	n := 0
	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading LigatureSet: "+"EOF: expected length: 2, got %d", L)
	}
	arrayLengthLigatures := int(binary.BigEndian.Uint16(src[0:]))
	n += 2

	{

		if L := len(src); L < 2+arrayLengthLigatures*2 {
			return item, 0, fmt.Errorf("reading LigatureSet: "+"EOF: expected length: %d, got %d", 2+arrayLengthLigatures*2, L)
		}

		item.Ligatures = make([]Ligature, arrayLengthLigatures) // allocation guarded by the previous check
		for i := range item.Ligatures {
			offset := int(binary.BigEndian.Uint16(src[2+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading LigatureSet: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.Ligatures[i], _, err = ParseLigature(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading LigatureSet: %s", err)
			}
		}
		n += arrayLengthLigatures * 2
	}
	return item, n, nil
}

func ParseLigatureSubs(src []byte) (LigatureSubs, int, error) {
	var item LigatureSubs
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading LigatureSubs: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.substFormat = binary.BigEndian.Uint16(src[0:])
	offsetCoverage := int(binary.BigEndian.Uint16(src[2:]))
	arrayLengthLigatureSets := int(binary.BigEndian.Uint16(src[4:]))
	n += 6

	{

		if offsetCoverage != 0 { // ignore null offset
			if L := len(src); L < offsetCoverage {
				return item, 0, fmt.Errorf("reading LigatureSubs: "+"EOF: expected length: %d, got %d", offsetCoverage, L)
			}

			var (
				err  error
				read int
			)
			item.Coverage, read, err = ParseCoverage(src[offsetCoverage:])
			if err != nil {
				return item, 0, fmt.Errorf("reading LigatureSubs: %s", err)
			}
			offsetCoverage += read
		}
	}
	{

		if L := len(src); L < 6+arrayLengthLigatureSets*2 {
			return item, 0, fmt.Errorf("reading LigatureSubs: "+"EOF: expected length: %d, got %d", 6+arrayLengthLigatureSets*2, L)
		}

		item.LigatureSets = make([]LigatureSet, arrayLengthLigatureSets) // allocation guarded by the previous check
		for i := range item.LigatureSets {
			offset := int(binary.BigEndian.Uint16(src[6+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading LigatureSubs: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.LigatureSets[i], _, err = ParseLigatureSet(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading LigatureSubs: %s", err)
			}
		}
		n += arrayLengthLigatureSets * 2
	}
	return item, n, nil
}

func ParseMultipleSubs(src []byte) (MultipleSubs, int, error) {
	var item MultipleSubs
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading MultipleSubs: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.substFormat = binary.BigEndian.Uint16(src[0:])
	offsetCoverage := int(binary.BigEndian.Uint16(src[2:]))
	arrayLengthSequences := int(binary.BigEndian.Uint16(src[4:]))
	n += 6

	{

		if offsetCoverage != 0 { // ignore null offset
			if L := len(src); L < offsetCoverage {
				return item, 0, fmt.Errorf("reading MultipleSubs: "+"EOF: expected length: %d, got %d", offsetCoverage, L)
			}

			var (
				err  error
				read int
			)
			item.Coverage, read, err = ParseCoverage(src[offsetCoverage:])
			if err != nil {
				return item, 0, fmt.Errorf("reading MultipleSubs: %s", err)
			}
			offsetCoverage += read
		}
	}
	{

		if L := len(src); L < 6+arrayLengthSequences*2 {
			return item, 0, fmt.Errorf("reading MultipleSubs: "+"EOF: expected length: %d, got %d", 6+arrayLengthSequences*2, L)
		}

		item.Sequences = make([]Sequence, arrayLengthSequences) // allocation guarded by the previous check
		for i := range item.Sequences {
			offset := int(binary.BigEndian.Uint16(src[6+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading MultipleSubs: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.Sequences[i], _, err = ParseSequence(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading MultipleSubs: %s", err)
			}
		}
		n += arrayLengthSequences * 2
	}
	return item, n, nil
}

func ParseReverseChainSingleSubs(src []byte) (ReverseChainSingleSubs, int, error) {
	var item ReverseChainSingleSubs
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.substFormat = binary.BigEndian.Uint16(src[0:])
	offsetCoverage := int(binary.BigEndian.Uint16(src[2:]))
	arrayLengthBacktrackCoverages := int(binary.BigEndian.Uint16(src[4:]))
	n += 6

	{

		if offsetCoverage != 0 { // ignore null offset
			if L := len(src); L < offsetCoverage {
				return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: "+"EOF: expected length: %d, got %d", offsetCoverage, L)
			}

			var (
				err  error
				read int
			)
			item.coverage, read, err = ParseCoverage(src[offsetCoverage:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: %s", err)
			}
			offsetCoverage += read
		}
	}
	{

		if L := len(src); L < 6+arrayLengthBacktrackCoverages*2 {
			return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: "+"EOF: expected length: %d, got %d", 6+arrayLengthBacktrackCoverages*2, L)
		}

		item.BacktrackCoverages = make([]Coverage, arrayLengthBacktrackCoverages) // allocation guarded by the previous check
		for i := range item.BacktrackCoverages {
			offset := int(binary.BigEndian.Uint16(src[6+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.BacktrackCoverages[i], _, err = ParseCoverage(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: %s", err)
			}
		}
		n += arrayLengthBacktrackCoverages * 2
	}
	if L := len(src); L < n+2 {
		return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: "+"EOF: expected length: n + 2, got %d", L)
	}
	arrayLengthLookaheadCoverages := int(binary.BigEndian.Uint16(src[n:]))
	n += 2

	{

		if L := len(src); L < n+arrayLengthLookaheadCoverages*2 {
			return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: "+"EOF: expected length: %d, got %d", n+arrayLengthLookaheadCoverages*2, L)
		}

		item.LookaheadCoverages = make([]Coverage, arrayLengthLookaheadCoverages) // allocation guarded by the previous check
		for i := range item.LookaheadCoverages {
			offset := int(binary.BigEndian.Uint16(src[n+i*2:]))
			// ignore null offsets
			if offset == 0 {
				continue
			}

			if L := len(src); L < offset {
				return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: "+"EOF: expected length: %d, got %d", offset, L)
			}

			var err error
			item.LookaheadCoverages[i], _, err = ParseCoverage(src[offset:])
			if err != nil {
				return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: %s", err)
			}
		}
		n += arrayLengthLookaheadCoverages * 2
	}
	if L := len(src); L < n+2 {
		return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: "+"EOF: expected length: n + 2, got %d", L)
	}
	arrayLengthSubstituteGlyphIDs := int(binary.BigEndian.Uint16(src[n:]))
	n += 2

	{

		if L := len(src); L < n+arrayLengthSubstituteGlyphIDs*2 {
			return item, 0, fmt.Errorf("reading ReverseChainSingleSubs: "+"EOF: expected length: %d, got %d", n+arrayLengthSubstituteGlyphIDs*2, L)
		}

		item.SubstituteGlyphIDs = make([]uint16, arrayLengthSubstituteGlyphIDs) // allocation guarded by the previous check
		for i := range item.SubstituteGlyphIDs {
			item.SubstituteGlyphIDs[i] = binary.BigEndian.Uint16(src[n+i*2:])
		}
		n += arrayLengthSubstituteGlyphIDs * 2
	}
	return item, n, nil
}

func ParseSequence(src []byte) (Sequence, int, error) {
	var item Sequence
	n := 0
	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading Sequence: "+"EOF: expected length: 2, got %d", L)
	}
	arrayLengthSubstituteGlyphIDs := int(binary.BigEndian.Uint16(src[0:]))
	n += 2

	{

		if L := len(src); L < 2+arrayLengthSubstituteGlyphIDs*2 {
			return item, 0, fmt.Errorf("reading Sequence: "+"EOF: expected length: %d, got %d", 2+arrayLengthSubstituteGlyphIDs*2, L)
		}

		item.SubstituteGlyphIDs = make([]uint16, arrayLengthSubstituteGlyphIDs) // allocation guarded by the previous check
		for i := range item.SubstituteGlyphIDs {
			item.SubstituteGlyphIDs[i] = binary.BigEndian.Uint16(src[2+i*2:])
		}
		n += arrayLengthSubstituteGlyphIDs * 2
	}
	return item, n, nil
}

func ParseSingleSubs(src []byte) (SingleSubs, int, error) {
	var item SingleSubs
	n := 0
	{
		var (
			err  error
			read int
		)
		item.Data, read, err = ParseSingleSubstData(src[0:])
		if err != nil {
			return item, 0, fmt.Errorf("reading SingleSubs: %s", err)
		}
		n += read
	}
	return item, n, nil
}

func ParseSingleSubstData(src []byte) (SingleSubstData, int, error) {
	var item SingleSubstData

	if L := len(src); L < 2 {
		return item, 0, fmt.Errorf("reading SingleSubstData: "+"EOF: expected length: 2, got %d", L)
	}
	format := uint16(binary.BigEndian.Uint16(src[0:]))
	var (
		read int
		err  error
	)
	switch format {
	case 1:
		item, read, err = ParseSingleSubstData1(src[0:])
	case 2:
		item, read, err = ParseSingleSubstData2(src[0:])
	default:
		err = fmt.Errorf("unsupported SingleSubstData format %d", format)
	}
	if err != nil {
		return item, 0, fmt.Errorf("reading SingleSubstData: %s", err)
	}

	return item, read, nil
}

func ParseSingleSubstData1(src []byte) (SingleSubstData1, int, error) {
	var item SingleSubstData1
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading SingleSubstData1: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.format = binary.BigEndian.Uint16(src[0:])
	offsetCoverage := int(binary.BigEndian.Uint16(src[2:]))
	item.DeltaGlyphID = int16(binary.BigEndian.Uint16(src[4:]))
	n += 6

	{

		if offsetCoverage != 0 { // ignore null offset
			if L := len(src); L < offsetCoverage {
				return item, 0, fmt.Errorf("reading SingleSubstData1: "+"EOF: expected length: %d, got %d", offsetCoverage, L)
			}

			var (
				err  error
				read int
			)
			item.Coverage, read, err = ParseCoverage(src[offsetCoverage:])
			if err != nil {
				return item, 0, fmt.Errorf("reading SingleSubstData1: %s", err)
			}
			offsetCoverage += read
		}
	}
	return item, n, nil
}

func ParseSingleSubstData2(src []byte) (SingleSubstData2, int, error) {
	var item SingleSubstData2
	n := 0
	if L := len(src); L < 6 {
		return item, 0, fmt.Errorf("reading SingleSubstData2: "+"EOF: expected length: 6, got %d", L)
	}
	_ = src[5] // early bound checking
	item.format = binary.BigEndian.Uint16(src[0:])
	offsetCoverage := int(binary.BigEndian.Uint16(src[2:]))
	arrayLengthSubstituteGlyphIDs := int(binary.BigEndian.Uint16(src[4:]))
	n += 6

	{

		if offsetCoverage != 0 { // ignore null offset
			if L := len(src); L < offsetCoverage {
				return item, 0, fmt.Errorf("reading SingleSubstData2: "+"EOF: expected length: %d, got %d", offsetCoverage, L)
			}

			var (
				err  error
				read int
			)
			item.Coverage, read, err = ParseCoverage(src[offsetCoverage:])
			if err != nil {
				return item, 0, fmt.Errorf("reading SingleSubstData2: %s", err)
			}
			offsetCoverage += read
		}
	}
	{

		if L := len(src); L < 6+arrayLengthSubstituteGlyphIDs*2 {
			return item, 0, fmt.Errorf("reading SingleSubstData2: "+"EOF: expected length: %d, got %d", 6+arrayLengthSubstituteGlyphIDs*2, L)
		}

		item.SubstituteGlyphIDs = make([]uint16, arrayLengthSubstituteGlyphIDs) // allocation guarded by the previous check
		for i := range item.SubstituteGlyphIDs {
			item.SubstituteGlyphIDs[i] = binary.BigEndian.Uint16(src[6+i*2:])
		}
		n += arrayLengthSubstituteGlyphIDs * 2
	}
	return item, n, nil
}
