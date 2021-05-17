// Copyright © 2019-2021 Talend - www.talend.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"unsafe"

	admv1 "k8s.io/api/admission/v1"
	admv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Note:
// =====
// These conversions come from https://github.com/jetstack/cert-manager/blob/ab0cd57dc58fd73a76fd96bd9d1402bd5ae96582/pkg/webhook/server/util/convert.go
// (which are adapted from https://github.com/kubernetes/kubernetes/blob/03d322035d2f199f2163658d94a153ed2b9de667/pkg/apis/admission/v1beta1/zz_generated.conversion.go)

func Convert_v1beta1_AdmissionReview_To_admission_AdmissionReview(in *admv1beta1.AdmissionReview, out *admv1.AdmissionReview) {
	if in.Request != nil {
		if out.Request == nil {
			out.Request = &admv1.AdmissionRequest{}
		}
		in, out := &in.Request, &out.Request
		*out = new(admv1.AdmissionRequest)
		Convert_v1beta1_AdmissionRequest_To_admission_AdmissionRequest(*in, *out)
	} else {
		out.Request = nil
	}
	out.Response = (*admv1.AdmissionResponse)(unsafe.Pointer(in.Response))
}

func Convert_v1beta1_AdmissionRequest_To_admission_AdmissionRequest(in *admv1beta1.AdmissionRequest, out *admv1.AdmissionRequest) {
	out.UID = types.UID(in.UID)
	out.Kind = in.Kind
	out.Resource = in.Resource
	out.SubResource = in.SubResource
	out.RequestKind = (*metav1.GroupVersionKind)(unsafe.Pointer(in.RequestKind))
	out.RequestResource = (*metav1.GroupVersionResource)(unsafe.Pointer(in.RequestResource))
	out.RequestSubResource = in.RequestSubResource
	out.Name = in.Name
	out.Namespace = in.Namespace
	out.Operation = admv1.Operation(in.Operation)
	out.Object = in.Object
	out.OldObject = in.OldObject
	out.Options = in.Options
}

func Convert_admission_AdmissionReview_To_v1beta1_AdmissionReview(in *admv1.AdmissionReview, out *admv1beta1.AdmissionReview) {
	if in.Request != nil {
		if out.Request == nil {
			out.Request = &admv1beta1.AdmissionRequest{}
		}
		in, out := &in.Request, &out.Request
		*out = new(admv1beta1.AdmissionRequest)
		Convert_admission_AdmissionRequest_To_v1beta1_AdmissionRequest(*in, *out)
	} else {
		out.Request = nil
	}
	out.Response = (*admv1beta1.AdmissionResponse)(unsafe.Pointer(in.Response))
}

func Convert_admission_AdmissionRequest_To_v1beta1_AdmissionRequest(in *admv1.AdmissionRequest, out *admv1beta1.AdmissionRequest) {
	out.UID = types.UID(in.UID)
	out.Kind = in.Kind
	out.Resource = in.Resource
	out.SubResource = in.SubResource
	out.RequestKind = (*metav1.GroupVersionKind)(unsafe.Pointer(in.RequestKind))
	out.RequestResource = (*metav1.GroupVersionResource)(unsafe.Pointer(in.RequestResource))
	out.RequestSubResource = in.RequestSubResource
	out.Name = in.Name
	out.Namespace = in.Namespace
	out.Operation = admv1beta1.Operation(in.Operation)
	out.Object = in.Object
	out.OldObject = in.OldObject
	out.Options = in.Options
}
