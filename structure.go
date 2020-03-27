package schematic

import (
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"reflect"
)

// schematic implements the structure of a Schematic, providing methods to read from it.
type schematic struct {
	Data      map[string]interface{}
	w, h, l   int
	materials string

	blocks   []uint8
	metadata []uint8
}

// init initialises the schematic structure, parsing several values from the NBT data.
func (s *schematic) init() error {
	s.w, s.h, s.l = int(s.Data["Width"].(int16)), int(s.Data["Height"].(int16)), int(s.Data["Length"].(int16))
	s.materials = s.Data["Materials"].(string)
	blocks, metadata := reflect.ValueOf(s.Data["Blocks"]), reflect.ValueOf(s.Data["Data"])
	blockSlice, metadataSlice := reflect.MakeSlice(reflect.SliceOf(blocks.Type().Elem()), blocks.Len(), blocks.Len()), reflect.MakeSlice(reflect.SliceOf(blocks.Type().Elem()), metadata.Len(), metadata.Len())
	reflect.Copy(blockSlice, blocks)
	reflect.Copy(metadataSlice, metadata)

	s.blocks, s.metadata = blockSlice.Interface().([]byte), metadataSlice.Interface().([]byte)
	if len(s.blocks) != s.w*s.h*s.l || len(s.metadata) != s.w*s.h*s.l {
		return fmt.Errorf("blocks and metadata were expected to be %v bytes long both (%v*%v*%v), but blocks has length %v and metadata length %v", s.w*s.h*s.l, s.w, s.h, s.l, len(s.blocks), len(s.metadata))
	}
	return nil
}

// Dimensions returns the dimensions of the schematic as width, height and length respectively.
func (s *schematic) Dimensions() [3]int {
	return [3]int{s.w, s.h, s.l}
}

// At returns the block found at a given position in the schematic. If any of the X, Y or Z coordinates passed
// are out of the bounds of the schematic, At will panic.
func (s *schematic) At(x, y, z int) world.Block {
	index := (y*s.l+z)*s.w + x
	id, meta := s.blocks[index], s.metadata[index]

	old := oldBlock{id: id, metadata: meta}
	if converted, ok := editionConversion[old]; ok {
		old = converted
	}

	n, ok := conversion[old]
	if !ok {
		return block.Air{}
	}
	ret := world_blocksHash[n.name+world_hashProperties(n.properties)]
	if ret == nil {
		return block.Air{}
	}
	return ret
}