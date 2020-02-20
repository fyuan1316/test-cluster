/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	productv1alpha1 "github.com/fyuan1316/test-cluster/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
)

// +kubebuilder:webhook:path=/mutate-product-alauda-io-v1alpha1-timatrix,mutating=true,failurePolicy=fail,groups=product.alauda.io,resources=timatrixes,verbs=create,versions=v1alpha1,name=mtmprovision.test.io

// +kubebuilder:webhook:verbs=create,path=/validate-product-alauda-io-v1alpha1-timatrix,mutating=false,failurePolicy=fail,groups=product.alauda.io,resources=timatrixes,versions=v1alpha1,name=vtmprovision.test.io

// log is for logging in this package.
var timatrixlog = logf.Log.WithName("timatrix-resource")

const (
	NotSupportMultipleTiMatrix = "NotSupportMultipleTiMatrix"
)

type MutatingWebHook struct {
	decoder *admission.Decoder
}

func (a *MutatingWebHook) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
func (a *MutatingWebHook) Handle(ctx context.Context, req admission.Request) admission.Response {
	ti := &productv1alpha1.TiMatrix{}
	err := a.decoder.Decode(req, ti)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	ti.Spec.Foo = "changed"
	allowedDeletable := true
	ti.Status.Deletable = &allowedDeletable
	marshaledPod, err := json.Marshal(ti)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

type ValidatingWebHook struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (a *ValidatingWebHook) ValidateCreate(ctx context.Context, req admission.Request) admission.Response {
	ti := &productv1alpha1.TiMatrix{}
	err := a.decoder.Decode(req, ti)
	if err != nil {
		timatrixlog.Error(err, "tmProvision webHook decode error")
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := validateTiSingleton(a.Client, ti); err != nil {
		timatrixlog.Error(err, "can not create more than one TiMatrix Platform in a Paas")
		return admission.Denied(NotSupportMultipleTiMatrix)
	}
	return admission.Allowed("")
}

func validateTiSingleton(dynClient client.Client, ti *productv1alpha1.TiMatrix) error {
	tiList := &productv1alpha1.TiMatrixList{}
	err := dynClient.List(context.Background(), tiList, &client.ListOptions{})
	if err == nil && len(tiList.Items) == 0 {
		return nil
	}
	if err == nil {
		err = fmt.Errorf("create %s failed, reason: %s", ti.Name, NotSupportMultipleTiMatrix)
	}
	return errors.NewInvalid(
		schema.GroupKind{Group: productv1alpha1.GroupVersion.Group, Kind: "tmprovision"},
		"", nil)
}

func (a *ValidatingWebHook) Handle(ctx context.Context, req admission.Request) admission.Response {
	fmt.Println("")
	if req.Operation == admissionv1beta1.Create {
		return a.ValidateCreate(ctx, req)
	}
	return admission.Allowed("")
}

func (a *ValidatingWebHook) InjectClient(c client.Client) error {
	a.Client = c
	return nil
}

func (a *ValidatingWebHook) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

func Register(mgr manager.Manager) {
	mgr.GetWebhookServer().Register(mPath, &webhook.Admission{Handler: &MutatingWebHook{}})
	mgr.GetWebhookServer().Register(vPath, &webhook.Admission{Handler: &ValidatingWebHook{Client: mgr.GetClient()}})
}

var (
	mPath = "/mutate-product-alauda-io-v1alpha1-timatrix"
	vPath = "/validate-product-alauda-io-v1alpha1-timatrix"
)
