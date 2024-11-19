import subprocess
import random
import time

# Config parameters
min_time = 5
max_time = 10
# TODO: add any missing non killable containers
non_killable_containers = ["projection4", "filter_english_7", "rabbitmq"]

output = subprocess.run(["docker", "container", "ls", "--format", "{{.Names}}"], capture_output=True)
containers = output.stdout.decode().splitlines()
killable_containers = [container for container in containers if container not in non_killable_containers]
gen_random_time = lambda: random.randint(min_time, max_time)

waiting_time = gen_random_time()
while True:
    container_to_kill = random.choice(killable_containers)
    subprocess.run(["docker", "container", "kill", container_to_kill], stdout=subprocess.DEVNULL)
    time.sleep(waiting_time)
    waiting_time = gen_random_time()
