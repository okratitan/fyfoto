/*
 * Copyright 2019 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bcgo

const (
	THRESHOLD_Z = 0
	THRESHOLD_I = 256 // (32 / 64) * 512 // Hash Space: 2 ^ 256 // Usage: General Purpose
	THRESHOLD_H = 272 // (34 / 64) * 512 // Hash Space: 2 ^ 240 // Usage: General Purpose
	THRESHOLD_G = 288 // (36 / 64) * 512 // Hash Space: 2 ^ 224 // Usage: General Purpose, Alias
	THRESHOLD_F = 304 // (38 / 64) * 512 // Hash Space: 2 ^ 208 // Usage: Hour Validation
	THRESHOLD_E = 320 // (40 / 64) * 512 // Hash Space: 2 ^ 192 // Usage: Day Validation
	THRESHOLD_D = 336 // (42 / 64) * 512 // Hash Space: 2 ^ 176 // Usage: Week Validation
	THRESHOLD_C = 352 // (44 / 64) * 512 // Hash Space: 2 ^ 160 // Usage: Year Validation
	THRESHOLD_B = 368 // (46 / 64) * 512 // Hash Space: 2 ^ 144 // Usage: Decade Validation
	THRESHOLD_A = 384 // (48 / 64) * 512 // Hash Space: 2 ^ 128 // Usage: Century Validation

	MAX_BLOCK_SIZE_BYTES   = uint64(2 * 1024 * 1024 * 1024) // 2Gb
	MAX_PAYLOAD_SIZE_BYTES = uint64(10 * 1024 * 1024)       // 10Mb
)
