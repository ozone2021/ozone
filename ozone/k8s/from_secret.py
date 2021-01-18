import os
import subprocess
import yaml

#from subprocess import check_output, CalledProcessError, STDOUT

secret_yaml_file = open(secret_filename)
parsed_yaml_file = yaml.load(secret_yaml_file, Loader=yaml.FullLoader)
secret_yaml_file.close()

env_vars = parsed_yaml_file["stringData"]

for env_var in env_vars:
    os.environ[env_var] = env_vars[env_var]

command = ""
for index, value in enumerate(sys.argv):
    if index >= start_index:
        command += value + " "

my_env = os.environ.copy()

process = subprocess.Popen(command, shell=True, env=my_env, stdout=subprocess.PIPE, universal_newlines=True)

# Poll process for new output until finished
while True:
    nextline = process.stdout.readline()
    if nextline == '' and process.poll() is not None:
        break
    sys.stdout.write(nextline)
    sys.stdout.flush()

output = process.communicate()[0]
exitCode = process.returncode

if (exitCode == 0):
    print(output)
else:
    raise ProcessException(command, exitCode, output)
# for line in proc.stdout:
#     print(line)
