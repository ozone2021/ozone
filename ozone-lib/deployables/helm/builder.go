package helm

import (
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"log"
	"os"
	"os/exec"
)

func getHelmParams() []string {
	return []string{
		"INSTALL_NAME",
		"FULL_TAG",
		"K8S_SERVICE",
		"CHART_DIR",
		"DOMAIN",
		//"GITLAB_PROJECT_CODE",
		//"BUILD_ARGS",
	}
}

func Deploy(serviceName string, env map[string]string) error {
	for _, arg := range getHelmParams() {
		if err := utils.ParamsOK("helmChart", arg, env); err != nil {
			return err
		}
	}

	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	installName := env["INSTALL_NAME"]
	chartDir := env["CHART_DIR"]
	k8sServiceName := env["K8S_SERVICE"]
	domain := env["DOMAIN"]
	subdomain := env["SUBDOMAIN"]
	tag := env["FULL_TAG"]

	containerPort, ok := env["CONTAINER_PORT"]
	if ok {
		utils.WarnIfNullVar(serviceName, containerPort,"CONTAINER_PORT")
		containerPort = fmt.Sprintf("--set service.containerPort=%s", containerPort)
	} else {
		containerPort = ""
	}

	servicePort, ok := env["SERVICE_PORT"]
	if ok {
		utils.WarnIfNullVar(serviceName, servicePort,"SERVICE_PORT")
		servicePort = fmt.Sprintf("--set service.servicePort=%s", servicePort)
	} else {
		servicePort = ""
	}

	namespace, ok := env["NAMESPACE"]
	if ok {
		namespace = fmt.Sprintf("-n %s --create-namespace", namespace)
	} else {
		namespace = ""
	}

	valuesFile, ok := env["VALUES_FILE"]
	if ok {
		valuesFile = fmt.Sprintf("-f %s", valuesFile)
	} else {
		valuesFile = ""
	}

	cmdString := fmt.Sprintf("helm upgrade --recreate-pods -i %s %s --set ingress.hosts[0].host=%s.%s%s --set image.fullTag=%s --set service.name=%s %s %s %s %s",
		installName,
		valuesFile,
		k8sServiceName,
		subdomain,
		domain,
		tag,
		k8sServiceName,
		containerPort,
		servicePort,
		namespace,
		chartDir,
	)

	log.Printf("Helm cmd is: %s", cmdString)

	cmdFields, argFields := process_manager.CommandFromFields(cmdString)
	cmd := exec.Command(cmdFields[0], argFields...)
	cmd.Dir = ozoneWorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("build docker err")
		return err
	}
	cmd.Wait()
	//query := &process_manager.ProcessCreateQuery{
	//	serviceName,
	//	ozoneWorkingDir,
	//	ozoneWorkingDir,
	//	cmdString,
	//	true,
	//	false,
	//	env,
	//}
	//
	//var reply *error
	//
	//client, err := rpc.DialHTTP("tcp", ":8000")
	//if err != nil {
	//	log.Fatal("dialing:", err)
	//}
	//err = client.Call("ProcessManager.AddProcess", query, reply)
	//if err != nil {
	//	log.Println(cmdString)
	//	log.Fatal("helm error:", err)
	//	return err
	//}
	return nil
}
