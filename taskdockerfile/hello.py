# Copyright 2020 Pradyumna Kaushik
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
import time
import signal
import sys

dev_null_file = open("/dev/null", "w")

# When we run 'docker stop', SIGTERM is sent to the running process.
def sig_handler(signal, frame):
    print("GOOD BYE FROM TASK-RANKER!!")
    dev_null_file.close()
    sys.exit(0)

signal.signal(signal.SIGTERM, sig_handler)
signal.signal(signal.SIGINT, sig_handler)

while True:
    f = open("/dev/null", "w")
    f.write("hello world")
    time.sleep(1)