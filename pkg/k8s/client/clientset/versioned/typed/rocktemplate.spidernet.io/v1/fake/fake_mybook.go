// Copyright 2022 Authors of spidernet-io
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	rocktemplatespidernetiov1 "github.com/spidernet-io/rocktemplate/pkg/k8s/apis/rocktemplate.spidernet.io/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMybooks implements MybookInterface
type FakeMybooks struct {
	Fake *FakeRocktemplateV1
}

var mybooksResource = schema.GroupVersionResource{Group: "rocktemplate.spidernet.io", Version: "v1", Resource: "mybooks"}

var mybooksKind = schema.GroupVersionKind{Group: "rocktemplate.spidernet.io", Version: "v1", Kind: "Mybook"}

// Get takes name of the mybook, and returns the corresponding mybook object, and an error if there is any.
func (c *FakeMybooks) Get(ctx context.Context, name string, options v1.GetOptions) (result *rocktemplatespidernetiov1.Mybook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(mybooksResource, name), &rocktemplatespidernetiov1.Mybook{})
	if obj == nil {
		return nil, err
	}
	return obj.(*rocktemplatespidernetiov1.Mybook), err
}

// List takes label and field selectors, and returns the list of Mybooks that match those selectors.
func (c *FakeMybooks) List(ctx context.Context, opts v1.ListOptions) (result *rocktemplatespidernetiov1.MybookList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(mybooksResource, mybooksKind, opts), &rocktemplatespidernetiov1.MybookList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &rocktemplatespidernetiov1.MybookList{ListMeta: obj.(*rocktemplatespidernetiov1.MybookList).ListMeta}
	for _, item := range obj.(*rocktemplatespidernetiov1.MybookList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mybooks.
func (c *FakeMybooks) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(mybooksResource, opts))
}

// Create takes the representation of a mybook and creates it.  Returns the server's representation of the mybook, and an error, if there is any.
func (c *FakeMybooks) Create(ctx context.Context, mybook *rocktemplatespidernetiov1.Mybook, opts v1.CreateOptions) (result *rocktemplatespidernetiov1.Mybook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(mybooksResource, mybook), &rocktemplatespidernetiov1.Mybook{})
	if obj == nil {
		return nil, err
	}
	return obj.(*rocktemplatespidernetiov1.Mybook), err
}

// Update takes the representation of a mybook and updates it. Returns the server's representation of the mybook, and an error, if there is any.
func (c *FakeMybooks) Update(ctx context.Context, mybook *rocktemplatespidernetiov1.Mybook, opts v1.UpdateOptions) (result *rocktemplatespidernetiov1.Mybook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(mybooksResource, mybook), &rocktemplatespidernetiov1.Mybook{})
	if obj == nil {
		return nil, err
	}
	return obj.(*rocktemplatespidernetiov1.Mybook), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMybooks) UpdateStatus(ctx context.Context, mybook *rocktemplatespidernetiov1.Mybook, opts v1.UpdateOptions) (*rocktemplatespidernetiov1.Mybook, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(mybooksResource, "status", mybook), &rocktemplatespidernetiov1.Mybook{})
	if obj == nil {
		return nil, err
	}
	return obj.(*rocktemplatespidernetiov1.Mybook), err
}

// Delete takes name of the mybook and deletes it. Returns an error if one occurs.
func (c *FakeMybooks) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(mybooksResource, name, opts), &rocktemplatespidernetiov1.Mybook{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMybooks) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(mybooksResource, listOpts)

	_, err := c.Fake.Invokes(action, &rocktemplatespidernetiov1.MybookList{})
	return err
}

// Patch applies the patch and returns the patched mybook.
func (c *FakeMybooks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *rocktemplatespidernetiov1.Mybook, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(mybooksResource, name, pt, data, subresources...), &rocktemplatespidernetiov1.Mybook{})
	if obj == nil {
		return nil, err
	}
	return obj.(*rocktemplatespidernetiov1.Mybook), err
}
