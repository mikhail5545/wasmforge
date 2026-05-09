/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

type (
	BucketType string

	ObjectRef struct {
		Bucket BucketType
		Key    string
	}

	ObjectInfo struct {
		Ref       ObjectRef
		SizeBytes int64
		Checksum  string
	}
)

const (
	BucketTypeArtifacts    BucketType = "artifacts"
	BucketTypeCertificates BucketType = "certificates"
	BucketTypeTmp          BucketType = "tmp"
)

func (t BucketType) String() string {
	return string(t)
}
