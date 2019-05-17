/*
 * LF: Global Fully Replicated Key/Value Store
 * Copyright (C) 2018-2019  ZeroTier, Inc.  https://www.zerotier.com/
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 *
 * --
 *
 * You can be released from the requirements of the license by purchasing
 * a commercial license. Buying such a license is mandatory as soon as you
 * develop commercial closed-source software that incorporates or links
 * directly against ZeroTier software without disclosing the source code
 * of your own application.
 */

package lf

import (
	"encoding/json"
	"net/http"
	"strings"
)

// APINewSelector (request, part of APINew) is a selector plain text name and an ordinal value (use zero if you don't care).
type APINewSelector struct {
	Name    Blob   `json:",omitempty"` // Name of this selector (masked so as to be hidden from those that don't know it)
	Ordinal uint64 `json:",omitempty"` // A sortable public value (optional)
}

// APINew (request) asks the proxy or node to perform server-side record generation and proof of work.
// Owners may be specified via raw OwnerPrivate key information or an OwnerSeed that is used server-side
// to deterministically (re-)generate the owner key pair. Note that both methods reveal your owner data.
// To avoid this generate the record locally and submit it directly instead of using the /new API.
type APINew struct {
	Selectors          []APINewSelector `json:",omitempty"` // Plain text selector names and ordinals
	MaskingKey         Blob             `json:",omitempty"` // Masking key to override default
	OwnerPrivate       Blob             `json:",omitempty"` // Full owner including private key (result of owner PrivateBytes() method)
	OwnerSeed          Blob             `json:",omitempty"` // Seed to deterministically generate owner (used if OwnerPrivate is missing)
	OwnerSeedOwnerType *byte            `json:",omitempty"` // Owner type for seeded owner mode (default: P-224)
	Links              []HashBlob       `json:",omitempty"` // Links to other records in the DAG
	Value              Blob             `json:",omitempty"` // Plain text (unmasked, uncompressed) value for this record
	Timestamp          *uint64          `json:",omitempty"` // Record timestamp in SECONDS since epoch (server time is used if zero or omitted)
}

// Run executes this API query against a remote LF node or proxy.
func (m *APINew) Run(url string) (*Record, error) {
	if strings.HasSuffix(url, "/") {
		url = url + "new"
	} else {
		url = url + "/new"
	}
	body, err := apiRun(url, m)
	if err != nil {
		return nil, err
	}
	var rec Record
	if err := json.Unmarshal(body, &rec); err != nil {
		return nil, err
	}
	return &rec, nil
}

func (m *APINew) execute(workFunction *Wharrgarblr) (*Record, *APIError) {
	var err error
	var owner *Owner
	if len(m.OwnerPrivate) > 0 {
		owner, err = NewOwnerFromPrivateBytes(m.OwnerPrivate)
		if err != nil {
			return nil, &APIError{Code: http.StatusBadRequest, Message: "cannot derive owner format public key from x509 private key: " + err.Error()}
		}
	} else if len(m.OwnerSeed) > 0 {
		ot := OwnerTypeNistP224
		if m.OwnerSeedOwnerType != nil {
			ot = *m.OwnerSeedOwnerType
		}
		owner, err = NewOwnerFromSeed(ot, m.OwnerSeed)
		if err != nil {
			return nil, &APIError{Code: http.StatusBadRequest, Message: "cannot generate owner from seed: " + err.Error()}
		}
	} else {
		return nil, &APIError{Code: http.StatusBadRequest, Message: "you must specify either 'ownerprivatekey' or 'ownerseed'"}
	}

	var ts uint64
	if m.Timestamp == nil || *m.Timestamp == 0 {
		ts = TimeSec()
	} else {
		ts = *m.Timestamp
	}

	sel := make([][]byte, len(m.Selectors))
	selord := make([]uint64, len(m.Selectors))
	for i := range m.Selectors {
		sel[i] = m.Selectors[i].Name
		selord[i] = m.Selectors[i].Ordinal
	}

	lnks := make([][32]byte, 0, len(m.Links))
	for _, l := range m.Links {
		lnks = append(lnks, l)
	}
	rec, err := NewRecord(RecordTypeDatum, m.Value, lnks, m.MaskingKey, sel, selord, nil, ts, workFunction, owner)
	if err != nil {
		return nil, &APIError{Code: http.StatusInternalServerError, Message: "record generation failed: " + err.Error()}
	}
	return rec, nil
}
