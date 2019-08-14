package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"strings"
)

var (
	getenvLong = `The kubectl-getenv plugin gets all the environment variables for the containers that run in the Pod.`
	listOfEnv  = make(map[string]string)
)

// GetEnvOptions provides information required to
// get the env from pods
type GetEnvOptions struct {
	configFlags *genericclioptions.ConfigFlags
	Pod         string
	Namespace   string
	args        []string
	ClientSet   *kubernetes.Clientset
	genericclioptions.IOStreams
}

// NewGetEnvOptions provides an instance of GetEnvOptions with default values
func NewGetEnvOptions(streams genericclioptions.IOStreams) *GetEnvOptions {
	return &GetEnvOptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

// NewCmdAks provides a cobra command wrapping AksOptions
func NewCmdGetenv(streams genericclioptions.IOStreams) *cobra.Command {

	o := NewGetEnvOptions(streams)

	cmd := &cobra.Command{
		Use:          "getenv",
		Short:        "Get environment variables for specific pod.",
		Long:         getenvLong,
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.GetEnv(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "default", "name of the namespace")
	return cmd
}

// Complete sets all required information for kubectl getenv plugin
func (o *GetEnvOptions) Complete(cmd *cobra.Command, args []string) error {

	var err error
	o.args = args

	if len(o.args) == 0 {
		return fmt.Errorf("You must specify the name of the pods to get the environment variables.")
	}

	if len(o.args) > 0 {
		o.Pod = args[0]
	}

	config, err := o.configFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	o.ClientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	return nil
}

// GetEnv gets all environment variable for specific container
func (o *GetEnvOptions) GetEnv() error {

	configMapEnv := make(map[string][]string)
	secretEnv := make(map[string][]string)

	pod, err := o.ClientSet.CoreV1().Pods(o.Namespace).Get(o.Pod, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for _, container := range pod.Spec.Containers {

		fmt.Printf("# %s\n\n", container.Name)
		for _, env := range container.Env {

			if env.ValueFrom == nil {

				// Considering other environment variables are plain text key-value pair
				// and also Skipping Kubnernetes specific ENV
				// For example: KUBERNETES_PORT_443_TCP_ADDR
				if !strings.HasPrefix(env.Name, "KUBERNETES_") {
					listOfEnv[env.Name] = env.Value
				}

			} else {

				switch {

				case env.ValueFrom.ConfigMapKeyRef != nil:

					configMapEnv[env.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name] = append(configMapEnv[env.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name], env.ValueFrom.ConfigMapKeyRef.Key)

				case env.ValueFrom.SecretKeyRef != nil:

					secretEnv[env.ValueFrom.SecretKeyRef.LocalObjectReference.Name] = append(secretEnv[env.ValueFrom.SecretKeyRef.LocalObjectReference.Name], env.ValueFrom.SecretKeyRef.Key)

				case env.ValueFrom.FieldRef != nil:

					listOfEnv[env.Name] = env.ValueFrom.FieldRef.FieldPath
				}
			}

		}

		err := o.getSecret(secretEnv)
		if err != nil {
			return err
		}
		err = o.getConfigMap(configMapEnv)
		if err != nil {
			return err
		}
		if len(listOfEnv) == 0 {
			fmt.Printf("There are no Environment Variables for the %s container.\n", container.Name)
		} else {
			for key, value := range listOfEnv {
				fmt.Printf("%s=%s\n", key, value)

				// Deleting keys from map once its printed
				delete(listOfEnv, key)
			}
		}
		fmt.Println()
	}
	return nil
}

// getSecret will fetch all secrets for provided secret
// and adds key-value pair on Global map (listOfEnv)
func (o *GetEnvOptions) getSecret(secretMap map[string][]string) error {

	for key, _ := range secretMap {
		secret, err := o.ClientSet.CoreV1().Secrets(o.Namespace).Get(key, metav1.GetOptions{})
		if err != nil {
			return err
		}
		for _, value := range secretMap[key] {

			listOfEnv[value] = string(secret.Data[value])
		}

	}
	return nil
}

// getConfigMap will fetch all configmap values for provided configMap
// and adds key-value pair Global map (listOfEnv)
func (o *GetEnvOptions) getConfigMap(configMap map[string][]string) error {

	for key, _ := range configMap {
		configmap, err := o.ClientSet.CoreV1().ConfigMaps(o.Namespace).Get(key, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for _, value := range configMap[key] {
			listOfEnv[value] = configmap.Data[value]
		}

	}
	return nil
}
