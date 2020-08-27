package controllers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/go-logr/logr"
	"github.com/raserge/configmap-operator/pkg/api/v1alpha1"
	"github.com/raserge/configmap-operator/third_party/kubernetes/forked/golang/expansion"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var pkgLog = log.Log.WithName("controllers")

// ConfigMapMake reconciles a ConfigMapMake object
type ConfigMapMake struct {
	client client.Client
	scheme *runtime.Scheme
	logger logr.Logger

	mu         sync.RWMutex
	configs    refMap
	configMaps refMap
	owned      refMap

	testNotifyFn func(types.NamespacedName)
}

// InjectClient injects the client into the reconciler.
func (r *ConfigMapMake) InjectClient(client client.Client) error {
	r.client = client
	return nil
}

// InjectLogger injects the logger into the reconciler.
func (r *ConfigMapMake) InjectLogger(logger logr.Logger) error {
	r.logger = logger.WithName("controller").WithName("ConfigMapMake")
	return nil
}

// InjectScheme injects the scheme into the reconciler.
func (r *ConfigMapMake) InjectScheme(scheme *runtime.Scheme) error {
	r.scheme = scheme
	return nil
}

// SetupWithManager sets up the reconciler with the manager.
func (r *ConfigMapMake) SetupWithManager(manager manager.Manager) error {
	if r.logger == nil {
		if err := r.InjectLogger(pkgLog); err != nil {
			return err
		}
	}

	return builder.ControllerManagedBy(manager).
		For(&v1alpha1.ConfigMapMake{}).
		Watches(&source.Kind{Type: &corev1.Secret{}}, handler.Funcs{
			CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
				r.configEventHandler(q, e.Object.(*corev1.Secret), false)
			},
			UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
				r.configEventHandler(q, e.ObjectNew.(*corev1.Secret), false)
			},
			DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
				r.configEventHandler(q, e.Object.(*corev1.Secret), true)
			},
			GenericFunc: func(e event.GenericEvent, q workqueue.RateLimitingInterface) {
				r.configEventHandler(q, e.Object.(*corev1.Secret), false)
			},
		}).
		Watches(&source.Kind{Type: &corev1.ConfigMap{}}, r.configMapEventHandler()).
		Complete(r)
}

func (r *ConfigMapMake) configMapEventHandler() handler.EventHandler {
	return &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(obj handler.MapObject) []reconcile.Request {
			namespace := obj.Meta.GetNamespace()
			name := obj.Meta.GetName()

			r.mu.RLock()
			defer r.mu.RUnlock()
			return toReqs(namespace, r.configMaps.srcs(namespace, name))
		}),
	}
}

func (r *ConfigMapMake) configEventHandler(q workqueue.RateLimitingInterface, config *corev1.Secret, deleted bool) {
	name := config.Name
	namespace := config.Namespace
	owner := getOwner(config)

	r.mu.Lock()
	if deleted || owner == nil {
		r.owned.set(namespace, name, nil)
	} else {
		r.owned.set(namespace, name, map[string]bool{string(owner.UID): true})
	}
	cmmNames := keys(r.configs.srcs(namespace, name))
	r.mu.Unlock()

	if owner != nil {
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      owner.Name,
			},
		})
	}
	for _, cmmName := range cmmNames {
		if owner != nil && owner.Name == cmmName {
			continue
		}
		q.Add(reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      cmmName,
			},
		})
	}
}

func (r *ConfigMapMake) setRefs(namespace, name string, configs, configMaps map[string]bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.configs.set(namespace, name, configs)
	r.configMaps.set(namespace, name, configMaps)
}

// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=configs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=configmaps.scartel.dc,resources=configmapmakes,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=configmaps.scartel.dc,resources=configmapmakes/status;configmapmakes/finalizers,verbs=get;update;patch

// Reconcile reconciles the state of the cluster with the desired state of a
// ConfigMapMake.
func (r *ConfigMapMake) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	if r.testNotifyFn != nil {
		defer r.testNotifyFn(req.NamespacedName)
	}
	ctx := context.TODO()
	log := r.logger.WithValues("configmapmake", req.NamespacedName)

	// Fetch the ConfigMapMake instance
	cmm := &v1alpha1.ConfigMapMake{}
	if err := r.client.Get(ctx, req.NamespacedName, cmm); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found. Owned objects are automatically garbage collected.
			r.setRefs(req.Namespace, req.Name, nil, nil)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	// Set the Secret and ConfigMap references for the instance
	configmapNames, configMapNames := varRefs(cmm.Spec.Vars)
	r.setRefs(cmm.Namespace, cmm.Name, configmapNames, configMapNames)

	// Sync and cleanup
	requeue, err := r.sync(ctx, log, cmm)
	if cleanupErr := r.cleanup(ctx, log, cmm); cleanupErr != nil && err == nil {
		err = cleanupErr
	}
	return reconcile.Result{Requeue: requeue}, err
}

func (r *ConfigMapMake) cleanup(ctx context.Context, log logr.Logger, cmm *v1alpha1.ConfigMapMake) error {
	configmapName := cmm.Spec.Template.Metadata.Name
	if configmapName == "" {
		configmapName = cmm.Name
	}

	r.mu.Lock()
	owned := keys(r.owned.srcs(cmm.Namespace, string(cmm.UID)))
	r.mu.Unlock()

	for _, name := range owned {
		if name == configmapName {
			continue
		}

		key := types.NamespacedName{Namespace: cmm.Namespace, Name: name}
		configLog := log.WithValues("config", key)
		configLog.Info("Cleaning up config")

		config := &corev1.ConfigMap{}
		if err := r.client.Get(ctx, key, config); err != nil {
			if apierrors.IsNotFound(err) {
				configLog.Info("Cleaning up config unnecessary, already removed")
				continue
			}
			configLog.Error(err, "Cleaning up config, get failed")
			return err
		}
		if err := r.client.Delete(ctx, config); err != nil {
			configLog.Error(err, "Cleaning up config, delete failed")
			return err
		}
	}
	return nil
}

func (r *ConfigMapMake) sync(ctx context.Context, log logr.Logger, cmm *v1alpha1.ConfigMapMake) (requeue bool, err error) {
	config, reason, err := r.renderConfig(ctx, cmm)
	if err != nil {
		msg := err.Error()
		defer func() {
			if statusErr := r.syncRenderFailureStatus(ctx, log, cmm, reason, msg); statusErr != nil {
				if err == nil {
					err = statusErr
				}
				requeue = true
			}
		}()
		if isConfigError(err) {
			log.Info("Unable to render ConfigMapMake", "warning", err)
			return true, nil
		}
		log.Error(err, "Unable to render ConfigMapMake")
		return false, err
	}

	key := types.NamespacedName{Namespace: config.Namespace, Name: config.Name}
	configLog := log.WithValues("config", key)

	// Check if the ConfigMap already exists
	found := &corev1.ConfigMap{}
	if err := r.client.Get(ctx, key, found); err != nil {
		if apierrors.IsNotFound(err) {
			configLog.Info("Creating Config")
			if err := r.client.Create(ctx, config); err != nil {
				configLog.Error(err, "Unable to create Config")
				return false, err
			}
			return false, r.syncSuccessStatus(ctx, log, cmm)
		}
		configLog.Error(err, "Unable to get Config")
		return false, err
	}

	// Confirm or take ownership.
	ownerChanged, err := r.setOwner(configLog, cmm, found)
	if err != nil {
		return false, err
	}

	// Update the object and write the result back if there are any changes
	if ownerChanged || shouldUpdate(found, config) {
		found.Labels = config.Labels
		found.Annotations = config.Annotations
		found.Data = config.Data
		// found.Type = config.Type
		configLog.Info("Updating Config")
		if err := r.client.Update(ctx, found); err != nil {
			configLog.Error(err, "Unable to update Config")
			return false, err
		}
	}
	return false, r.syncSuccessStatus(ctx, log, cmm)
}

func (r *ConfigMapMake) setOwner(log logr.Logger, cmm *v1alpha1.ConfigMapMake, config *corev1.ConfigMap) (bool, error) {
	gvk, err := apiutil.GVKForObject(cmm, r.scheme)
	if err != nil {
		return false, err
	}
	owner := metav1.NewControllerRef(cmm, gvk)
	for i, ref := range config.OwnerReferences {
		if ref.Controller == nil || !*ref.Controller {
			continue
		}
		if ref.UID != cmm.UID {
			log.Error(err, "Config has a different owner", "owner", ref)
			return false, &controllerutil.AlreadyOwnedError{Object: cmm, Owner: ref}
		}
		if !reflect.DeepEqual(&ref, owner) { // e.g. apiVersion changed
			log.Info("Updating ownership of Config")
			config.OwnerReferences[i] = *owner
			return true, nil
		}
		return false, nil
	}
	log.Info("Taking ownership of Config", "owner", *owner)
	cmm.OwnerReferences = append(config.OwnerReferences, *owner)
	return true, nil
}

func shouldUpdate(a, b *corev1.ConfigMap) bool {
	// return a.Type != b.Type ||
	return !reflect.DeepEqual(a.Annotations, b.Annotations) ||
		!reflect.DeepEqual(a.Labels, b.Labels) ||
		!reflect.DeepEqual(a.Data, b.Data)
}

func (r *ConfigMapMake) renderConfig(ctx context.Context, cmm *v1alpha1.ConfigMapMake) (*corev1.ConfigMap, string, error) {
	vars, err := r.makeVariables(ctx, cmm)
	if err != nil {
		return nil, CreateVariablesErrorReason, err
	}
	varMapFn := expansion.MappingFuncFor(vars)

	data := make(map[string]string)
	binarydata := make(map[string][]byte)
	for k, v := range cmm.Spec.Template.Data {
		data[k] = string(expansion.Expand(v, varMapFn))
	}
	for k, v := range cmm.Spec.Template.BinaryData {
		binarydata[k] = []byte(expansion.Expand(string(v), varMapFn))
	}

	meta := cmm.Spec.Template.Metadata
	name := meta.Name
	if name == "" {
		name = cmm.Name
	}
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   cmm.Namespace,
			Labels:      meta.Labels,
			Annotations: meta.Annotations,
		},
		BinaryData: binarydata,
		Data:       data,
	}
	if err := controllerutil.SetControllerReference(cmm, config, r.scheme); err != nil {
		return nil, internalError, err
	}
	return config, "", nil
}

// Same logic as container env vars: Kubelet.makeEnvironmentVariables
// https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kubelet_pods.go
func (r *ConfigMapMake) makeVariables(ctx context.Context, cmm *v1alpha1.ConfigMapMake) (vars map[string]string, err error) {
	vars = make(map[string]string)
	mappingFn := expansion.MappingFuncFor(vars)
	configMaps := make(map[string]*corev1.ConfigMap)
	secrets := make(map[string]*corev1.Secret)

	for _, v := range cmm.Spec.Vars {
		val := v.Value
		found := true

		switch {
		case val != "":
			val = expansion.Expand(val, mappingFn)
		case v.SecretValue != nil:
			val, found, err = r.configValue(ctx, secrets, cmm.Namespace, v.SecretValue)
		case v.ConfigMapValue != nil:
			val, found, err = r.configMapValue(ctx, configMaps, cmm.Namespace, v.ConfigMapValue)
		}

		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}

		vars[v.Name] = val
	}

	return vars, nil
}

func (r *ConfigMapMake) configValue(ctx context.Context, cache map[string]*corev1.Secret, namespace string, ref *corev1.SecretKeySelector) (value string, found bool, err error) {
	name := ref.Name
	key := ref.Key
	optional := ref.Optional != nil && *ref.Optional

	config, found := cache[name]
	if !found {
		config = &corev1.Secret{}
		err := r.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, config)
		if err != nil {
			if apierrors.IsNotFound(err) {
				if optional {
					return "", false, nil
				}
				return "", false, &configError{err}
			}
			return "", false, err
		}
		cache[name] = config
	}
	if buf, found := config.Data[key]; found {
		return string(buf), true, nil
	}
	if optional {
		return "", false, nil
	}
	return "", false, newConfigError("Couldn't find key %s in Secret %s/%s", key, namespace, name)
}

func (r *ConfigMapMake) configMapValue(ctx context.Context, cache map[string]*corev1.ConfigMap, namespace string, ref *corev1.ConfigMapKeySelector) (value string, found bool, err error) {
	name := ref.Name
	key := ref.Key
	optional := ref.Optional != nil && *ref.Optional

	configMap, found := cache[name]
	if !found {
		configMap = &corev1.ConfigMap{}
		err := r.client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, configMap)
		if err != nil {
			if apierrors.IsNotFound(err) {
				if optional {
					return "", false, nil
				}
				return "", false, &configError{err}
			}
			return "", false, err
		}
		cache[name] = configMap
	}
	if str, found := configMap.Data[key]; found {
		return str, true, nil
	}
	if buf, found := configMap.BinaryData[key]; found {
		return string(buf), true, nil
	}
	if optional {
		return "", false, nil
	}
	return "", false, newConfigError("Couldn't find key %s in ConfigMap %s/%s", key, namespace, name)
}

func (r *ConfigMapMake) syncSuccessStatus(ctx context.Context, log logr.Logger, cmm *v1alpha1.ConfigMapMake) error {
	return r.syncStatus(ctx, log, cmm, corev1.ConditionFalse, "", "")
}

func (r *ConfigMapMake) syncRenderFailureStatus(ctx context.Context, log logr.Logger, cmm *v1alpha1.ConfigMapMake, reason, message string) error {
	return r.syncStatus(ctx, log, cmm, corev1.ConditionTrue, reason, message)
}

func (r *ConfigMapMake) syncStatus(ctx context.Context, log logr.Logger, cmm *v1alpha1.ConfigMapMake, condStatus corev1.ConditionStatus, reason, message string) error {
	status := v1alpha1.ConfigMapMakeStatus{
		ObservedGeneration: cmm.Generation,
		Conditions:         cmm.Status.Conditions,
	}
	cond := NewConfigMapMakeCondition(v1alpha1.ConfigMapMakeRenderFailure, condStatus, reason, message)
	SetConfigMapMakeCondition(&status, *cond) // original backing array not modified
	if reflect.DeepEqual(cmm.Status, status) {
		return nil
	}
	cmm.Status = status
	log.Info("Updating status")
	if err := r.client.Status().Update(ctx, cmm); err != nil {
		log.Error(err, "Unable to update status")
		return err
	}
	return nil
}

func getOwner(config *corev1.Secret) *metav1.OwnerReference {
	owner := metav1.GetControllerOf(config)
	if owner == nil || owner.Kind != "ConfigMapMake" {
		return nil
	}
	if gv, _ := schema.ParseGroupVersion(owner.APIVersion); gv.Group != v1alpha1.GroupVersion.Group {
		return nil
	}
	return owner
}

func keys(set map[string]bool) []string {
	n := len(set)
	if n == 0 {
		return nil
	}
	s := make([]string, 0, n)
	for k := range set {
		s = append(s, k)
	}
	return s
}

func toReqs(namespace string, names map[string]bool) []reconcile.Request {
	var reqs []reconcile.Request
	for name := range names {
		reqs = append(reqs, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      name,
			},
		})
	}
	return reqs
}

func varRefs(vars []v1alpha1.TemplateVariable) (configs, configMaps map[string]bool) {
	for _, v := range vars {
		if v.SecretValue != nil {
			if configs == nil {
				configs = make(map[string]bool)
			}
			configs[v.SecretValue.Name] = true
		}
		if v.ConfigMapValue != nil {
			if configMaps == nil {
				configMaps = make(map[string]bool)
			}
			configMaps[v.ConfigMapValue.Name] = true
		}
	}
	return configs, configMaps
}

type configError struct {
	err error
}

func newConfigError(format string, v ...interface{}) *configError {
	if len(v) == 0 {
		return &configError{errors.New(format)}
	}
	return &configError{fmt.Errorf(format, v...)}
}

func (e *configError) Error() string {
	return e.err.Error()
}

func (*configError) IsConfigError() bool { return true }

func isConfigError(err error) bool {
	v, ok := err.(interface {
		IsConfigError() bool
	})
	return ok && v.IsConfigError()
}
