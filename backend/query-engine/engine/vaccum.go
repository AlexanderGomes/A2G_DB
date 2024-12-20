package engine

import (
	"a2gdb/storage-engine/storage"
	"log"
)

const (
	ALMOST_FULL_PAGE = 486
	EIGHT_LEVEL      = 886
	SEVENTH_LEVEL    = 1286
	SIXTH_LEVEL      = 1686
	FIFITH_LEVEL     = 2086
	FOURTH_LEVEL     = 2486
	THIRD_LEVEL      = 2886
	SECOND_LEVEL     = 3286
	FIRST_LEVEL      = 3686
	EMPTY_PAGE       = 4082
	NEXT_LEVEL       = 400
)

func vaccumEntry(newSpace []*storage.FreeSpace, tableObj *storage.TableObj) {
	claimCompressSpace(newSpace, tableObj)
	memSeparationMass(newSpace, tableObj)
}

func claimCompressSpace(newSpace []*storage.FreeSpace, tableObj *storage.TableObj) {
	for _, page := range newSpace {
		newPage, err := storage.RearrangePAGE(page.TempPagePtr, tableObj)
		if err != nil {
			log.Fatalf("failed rearrage page %+v, error: %s", page, err)
		}

		err = updatePageInfo(nil, newPage, tableObj)
		if err != nil {
			log.Fatalf("failed updaing page, error: %s", err)
		}
	}
}

func memSeparationMass(newSpace []*storage.FreeSpace, tableObj *storage.TableObj) {
	for _, space := range newSpace {
		// find page
		// remove from current league
		// add to new league
		
		memTag := getTag(space.FreeMemory)
		tableObj.Memory[memTag] = append(tableObj.Memory[memTag], space)
	}

	saveMemMapping(tableObj)
}

func memSeparationSingle(newSpace storage.FreeSpace, tableObj *storage.TableObj) {
	memTag := getTag(newSpace.FreeMemory)
	tableObj.Memory[memTag] = append(tableObj.Memory[memTag], &newSpace)
	saveMemMapping(tableObj)
}

func saveMemMapping(tableObj *storage.TableObj) {
	bytes, err := storage.EncodeMemObj(tableObj.Memory)
	if err != nil {
		log.Fatalf("encoding mem obj failed: %s", err)
	}

	err = storage.WriteNonPageFile(tableObj.MemFile, bytes)
	if err != nil {
		log.Fatalf("writing mem obj failed: %s", err)
	}
}

func searchPage(tableObj *storage.TableObj, memoryNedded, level uint16) ([]*storage.FreeSpace, *storage.FreeSpace, uint16, int) {
	var memSlice []*storage.FreeSpace
	var memTag uint16
	var spaceInfo *storage.FreeSpace
	var deleteIndex int

	for len(memSlice) == 0 {
		memTag = getTag(level)

		//# create page
		if memTag == 0 {
			return nil, nil, memTag, 0
		}

		memSlice = tableObj.Memory[memTag]
		//# empty level
		if len(memSlice) == 0 {
			level += NEXT_LEVEL
			continue
		}

		for i, mem := range memSlice {
			if memoryNedded <= mem.FreeMemory {
				spaceInfo = mem
				deleteIndex = i
				break
			}
		}

		//# free space not found
		if spaceInfo == nil {
			memSlice = []*storage.FreeSpace{}
			level += NEXT_LEVEL
			continue
		}

	}

	return memSlice, spaceInfo, memTag, deleteIndex
}

func getAvailablePage(tableObj *storage.TableObj, memoryNedded uint16) *storage.PageV2 {
	var memSlice []*storage.FreeSpace
	var memTag uint16
	var availablePage *storage.PageV2
	var spaceInfo *storage.FreeSpace
	var deleteIndex int

	memSlice, spaceInfo, memTag, deleteIndex = searchPage(tableObj, memoryNedded, memoryNedded)
	if memTag == 0 {
		return storage.CreatePageV2()
	}

	memSlice = append(memSlice[:deleteIndex], memSlice[deleteIndex+1:]...)
	tableObj.Memory[memTag] = memSlice

	pageInfo := tableObj.DirectoryPage.Value[spaceInfo.PageID]
	pageBytes, err := storage.ReadPageAtOffset(tableObj.DataFile, pageInfo.Offset)
	if err != nil {
		log.Panicf("reading page at offset %s failed", err)
	}

	availablePage, err = storage.DecodePageV2(pageBytes)
	if err != nil {
		log.Panicf("decoding page from offset %s failed", err)
	}

	return availablePage
}

func getTag(pageMem uint16) uint16 {
	switch {
	case pageMem <= ALMOST_FULL_PAGE:
		return ALMOST_FULL_PAGE
	case pageMem > SECOND_LEVEL && pageMem <= FIRST_LEVEL:
		return FIRST_LEVEL
	case pageMem > THIRD_LEVEL && pageMem <= SECOND_LEVEL:
		return SECOND_LEVEL
	case pageMem > FOURTH_LEVEL && pageMem <= THIRD_LEVEL:
		return THIRD_LEVEL
	case pageMem > FIFITH_LEVEL && pageMem <= FOURTH_LEVEL:
		return FOURTH_LEVEL
	case pageMem > SIXTH_LEVEL && pageMem <= FIFITH_LEVEL:
		return FIFITH_LEVEL
	case pageMem > SEVENTH_LEVEL && pageMem <= SIXTH_LEVEL:
		return SIXTH_LEVEL
	case pageMem > EIGHT_LEVEL && pageMem <= SEVENTH_LEVEL:
		return SEVENTH_LEVEL
	case pageMem > ALMOST_FULL_PAGE && pageMem <= EIGHT_LEVEL:
		return EIGHT_LEVEL
	case pageMem > FIRST_LEVEL && pageMem <= EMPTY_PAGE:
		return EMPTY_PAGE
	default:
		return 0
	}
}
