//go:build linux || darwin || windows

package ffi3

/*
#include <stddef.h> // For size_t
#include <stdlib.h> // For malloc, free

// A struct to pass type information (size and alignment) from Go to C.
typedef struct {
    size_t size;
    size_t alignment;
} type_info_t;

// A struct to pass the calculated layout information from C back to Go.
typedef struct {
    size_t total_size;
    size_t total_alignment;
    size_t* offsets;
} layout_info_t;

// This C function is the brain of the layout calculation. It mimics how a C
// compiler would lay out a struct in memory, respecting alignment and padding.
layout_info_t calculate_struct_layout(type_info_t* fields, int num_fields) {
    layout_info_t layout;
    layout.offsets = NULL;
    layout.total_size = 0;
    layout.total_alignment = 0;

    if (num_fields <= 0) {
        return layout;
    }

    layout.offsets = (size_t*)malloc(sizeof(size_t) * num_fields);
    if (layout.offsets == NULL) {
        // Return a zeroed struct to indicate failure. Go side should check for NULL.
        return layout;
    }

    size_t current_offset = 0;
    size_t max_alignment = 0;

    for (int i = 0; i < num_fields; i++) {
        size_t alignment = fields[i].alignment;
        if (alignment == 0) continue; // Skip zero-sized fields

        // Update the struct's overall alignment requirement.
        if (alignment > max_alignment) {
            max_alignment = alignment;
        }

        // Add padding to satisfy the current field's alignment.
        // This is the key step: ensure the offset is a multiple of the alignment.
        size_t padding = (alignment - (current_offset % alignment)) % alignment;
        current_offset += padding;

        // Store the calculated offset for this field.
        layout.offsets[i] = current_offset;

        // Advance the offset by the size of the current field.
        current_offset += fields[i].size;
    }

    // Add final padding to the end of the struct so its total size is a
    // multiple of its largest alignment. This is crucial for arrays of structs.
    if (max_alignment > 0) {
        size_t final_padding = (max_alignment - (current_offset % max_alignment)) % max_alignment;
        current_offset += final_padding;
    }

    layout.total_size = current_offset;
    layout.total_alignment = max_alignment;

    return layout;
}

// Helper to free the memory allocated inside the layout_info_t struct.
void free_layout_info(layout_info_t layout) {
    if (layout.offsets != NULL) {
        free(layout.offsets);
    }
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// calculateLayout is the Go wrapper around the C helper function. It bridges the
// gap between Go's []StructField and C's type_info_t array.
func calculateLayout(fields []StructField) (totalSize, totalAlignment uintptr, offsets []uintptr, err error) {
	numFields := len(fields)
	if numFields == 0 {
		return 0, 0, nil, nil
	}

	// Allocate C memory for the input array.
	cTypeInfos := (*C.type_info_t)(C.malloc(C.size_t(numFields) * C.size_t(unsafe.Sizeof(C.type_info_t{}))))
	if cTypeInfos == nil {
		return 0, 0, nil, fmt.Errorf("malloc failed for type_info array")
	}
	defer C.free(unsafe.Pointer(cTypeInfos))

	// Convert the Go slice into a C array view.
	infoSlice := (*[1 << 30]C.type_info_t)(unsafe.Pointer(cTypeInfos))[:numFields:numFields]
	for i, field := range fields {
		infoSlice[i].size = C.size_t(field.Type.Size())
		infoSlice[i].alignment = C.size_t(field.Type.Alignment())
	}

	// Call the C function to perform the calculation.
	cLayout := C.calculate_struct_layout(cTypeInfos, C.int(numFields))
	if cLayout.offsets == nil {
		return 0, 0, nil, fmt.Errorf("calculate_struct_layout failed, likely out of memory")
	}
	defer C.free_layout_info(cLayout)

	// Convert the results from C types back to Go types.
	totalSize = uintptr(cLayout.total_size)
	totalAlignment = uintptr(cLayout.total_alignment)
	offsets = make([]uintptr, numFields)

	// Convert the C offsets array into a Go slice view.
	cOffsetsSlice := (*[1 << 30]C.size_t)(unsafe.Pointer(cLayout.offsets))[:numFields:numFields]
	for i := 0; i < numFields; i++ {
		offsets[i] = uintptr(cOffsetsSlice[i])
	}

	return totalSize, totalAlignment, offsets, nil
}