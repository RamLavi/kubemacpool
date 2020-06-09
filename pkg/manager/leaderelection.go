package manager

import (
	"context"
	"github.com/k8snetworkplumbingwg/kubemacpool/pkg/names"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *KubeMacPoolManager) waitToStartLeading() error {
	<-k.mgr.Elected()
	// If we reach here then we are in the elected pod.
	return k.markPodAsLeader()
}

func (k *KubeMacPoolManager) markPodAsLeader() error {
	pod, err := k.clientset.CoreV1().Pods(k.podNamespace).Get(context.TODO(), k.podName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	pod.Labels[names.LEADER_LABEL] = "true"
	_, err = k.clientset.CoreV1().Pods(k.podNamespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	log.Info("marked this manager as leader for webhook", "podName", k.podName)
	return nil
}
