/**
 * Copyright (C) 2015 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp

import (
	"encoding/binary"
	"net"
	"runtime"

	"golang.org/x/net/internal/socket"
	"golang.org/x/net/ipv4"
)

// freebsdVersion is set in sys_freebsd.go.
// See http://www.freebsd.org/doc/en/books/porters-handbook/freebsd-versions.html.
var freebsdVersion uint32

// ParseIPv4Header parses b as an IPv4 header of ICMP error message
// invoking packet, which is contained in ICMP error message.
func ParseIPv4Header(b []byte) (*ipv4.Header, error) {
	if len(b) < ipv4.HeaderLen {
		return nil, errHeaderTooShort
	}
	hdrlen := int(b[0]&0x0f) << 2
	if hdrlen > len(b) {
		return nil, errBufferTooShort
	}
	h := &ipv4.Header{
		Version:  int(b[0] >> 4),
		Len:      hdrlen,
		TOS:      int(b[1]),
		ID:       int(binary.BigEndian.Uint16(b[4:6])),
		FragOff:  int(binary.BigEndian.Uint16(b[6:8])),
		TTL:      int(b[8]),
		Protocol: int(b[9]),
		Checksum: int(binary.BigEndian.Uint16(b[10:12])),
		Src:      net.IPv4(b[12], b[13], b[14], b[15]),
		Dst:      net.IPv4(b[16], b[17], b[18], b[19]),
	}
	switch runtime.GOOS {
	case "darwin":
		h.TotalLen = int(socket.NativeEndian.Uint16(b[2:4]))
	case "freebsd":
		if freebsdVersion >= 1000000 {
			h.TotalLen = int(binary.BigEndian.Uint16(b[2:4]))
		} else {
			h.TotalLen = int(socket.NativeEndian.Uint16(b[2:4]))
		}
	default:
		h.TotalLen = int(binary.BigEndian.Uint16(b[2:4]))
	}
	h.Flags = ipv4.HeaderFlags(h.FragOff&0xe000) >> 13
	h.FragOff = h.FragOff & 0x1fff
	if hdrlen-ipv4.HeaderLen > 0 {
		h.Options = make([]byte, hdrlen-ipv4.HeaderLen)
		copy(h.Options, b[ipv4.HeaderLen:])
	}
	return h, nil
}