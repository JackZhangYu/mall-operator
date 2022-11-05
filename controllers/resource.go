package controllers

import (
	"context"
	"fmt"
	v1 "mall-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	MallWebCommonLabelkey = "app"
)

const (
	APP_NAME = "mall-app"

	CONTAINER_PORT = 80

	CPU_REQUEST = "100m"

	CPU_LIMIT = "100m"

	MEM_REQUEST = "50Mi"

	MEM_LIMIT = "50Mi"
)

func getExpectReplicas(mallWeb *v1.MallWeb) int32 {
	singlePodQPS := *mallWeb.Spec.SinglePodsQPS

	totalQPS := *mallWeb.Spec.TotalQPS

	replicas := totalQPS / singlePodQPS

	if totalQPS%singlePodQPS != 0 {
		replicas += 1
	}

	return replicas
}

// If not exist service,create it.
func CreateServiceIfNotExists(ctx context.Context, r *MallWebReconciler, mallWeb *v1.MallWeb, req ctrl.Request) error {
	logger := log.FromContext(ctx)

	logger.WithValues("func", "CreateService")
	svc := &corev1.Service{}

	svc.Name = mallWeb.Name
	svc.Namespace = mallWeb.Namespace
	svc.Spec = corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{
				Name:     "http",
				NodePort: *mallWeb.Spec.Port,
				Port:     int32(CONTAINER_PORT),
			},
		},
		Type: corev1.ServiceTypeNodePort,
		Selector: map[string]string{
			MallWebCommonLabelkey: APP_NAME,
		},
	}

	//set connect relation
	logger.Info("set reference")

	if err := controllerutil.SetControllerReference(mallWeb, svc, r.Scheme); err != nil {
		logger.Error(err, "set controller reference failed")
		return err
	}
	logger.Info("start create service")
	if err := r.Create(ctx, svc); err != nil {
		logger.Error(err, "create service error")
		return err
	}

	return nil
}

// CreateDeployment create the deployment
func CreateDeployment(ctx context.Context, r *MallWebReconciler, mallWeb *v1.MallWeb) error {

	logger := log.FromContext(ctx)
	logger.WithValues("func", "createDeploy")

	expectReplicas := getExpectReplicas(mallWeb)

	logger.Info(fmt.Sprintf("expectReplicas [%d]", expectReplicas))

	deploy := &appsv1.Deployment{}

	deploy.Labels = map[string]string{
		MallWebCommonLabelkey: APP_NAME,
	}

	deploy.Name = mallWeb.Name
	deploy.Namespace = mallWeb.Namespace

	deploy.Spec = appsv1.DeploymentSpec{
		Replicas: pointer.Int32Ptr(expectReplicas),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				MallWebCommonLabelkey: APP_NAME,
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					MallWebCommonLabelkey: APP_NAME,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  APP_NAME,
						Image: mallWeb.Spec.Image,
						Ports: []corev1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: CONTAINER_PORT,
								Protocol:      corev1.ProtocolSCTP,
							},
						},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(CPU_LIMIT),
								corev1.ResourceMemory: resource.MustParse(MEM_LIMIT),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(CPU_REQUEST),
								corev1.ResourceMemory: resource.MustParse(MEM_REQUEST),
							},
						},
					},
				},
			},
		},
	}
	// build reference ,delete web and will delete the deploy
	logger.Info("set reference")
	if err := controllerutil.SetControllerReference(mallWeb, deploy, r.Scheme); err != nil {
		logger.Error(err, "set controller reference error")
		return err
	}
	// create deployment
	logger.Info("start create deployment")
	if err := r.Create(ctx, deploy); err != nil {
		logger.Error(err, "create deployment error")
		return err
	}
	logger.Info("create deploy success")
	return nil

}

func updateStatus(ctx context.Context, r *MallWebReconciler, mallWeb *v1.MallWeb) error {
	logger := log.FromContext(ctx)
	logger.WithValues("func", "updateStatus")
	//single pod's QPS
	singlePodQPS := *mallWeb.Spec.SinglePodsQPS
	// pod replicas
	replicas := getExpectReplicas(mallWeb)

	if nil == mallWeb.Status.RealQPS {
		mallWeb.Status.RealQPS = new(int32)
	}

	*mallWeb.Status.RealQPS = singlePodQPS * replicas
	logger.Info(fmt.Sprintf("singlePodQPS [%d], replicas [%d],realQPS[%d]", singlePodQPS, replicas, *&mallWeb.Status.RealQPS))

	if err := r.Update(ctx, mallWeb); err != nil {
		logger.Error(err, "update instance error")
		return err
	}
	return nil

}
