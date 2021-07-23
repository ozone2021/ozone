package utilities

import (
    "fmt"
    "github.com/ozone2021/ozone/ozone-lib/utils"
    "os"
    "os/exec"
)

func getParams() []string {
    return []string{
        "SCRIPT",
    }
}

func RunBashScript(env map[string]string) error {
    for _, arg := range getParams() {
        if err := utils.ParamsOK("RunBashScript", arg, env); err != nil {
            return err
        }
    }
    scriptPath := env["SCRIPT"]
    cmd := exec.Command("/bin/bash", scriptPath)
    cmd.Env = os.Environ()
    for k, v := range env {
        cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
    }
    output, err := cmd.Output()

    fmt.Println(string(output))

    if err != nil {
        return err
    }
    //output := string(out)
    return nil
}
