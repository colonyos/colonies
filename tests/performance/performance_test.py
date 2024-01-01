from pycolonies import func_spec
from pycolonies import colonies_client
from pycolonies import Crypto
import csv
import os
from datetime import datetime, timedelta
import numpy as np
import time

def submit(colonies, colonyname, prvkey):
    f = func_spec(func="helloworld", 
                  args=[], 
                  colonyname=colonyname, 
                  executortype="test-executor",
                  priority=200,
                  maxexectime=-1,
                  maxretries=-1,
                  maxwaittime=-1)
    colonies.submit(f, prvkey)
   
def assign(colonies, colonyname, prvkey):
    process = colonies.assign(colonyname, 10, prvkey)
    colonies.close(process["processid"], ["helloworld"], prvkey)

def register(colonies, colonyname, colony_prvkey):
    crypto = Crypto()
    executor_prvkey = crypto.prvkey()
    executorid = crypto.id(executor_prvkey)

    executor = {
        "executorname": "test-executor",
        "executorid": executorid,
        "colonyname": colonyname,
        "executortype": "test-executor"
    }
    
    try:
        executor = colonies.add_executor(executor, colony_prvkey)
        colonies.approve_executor(colonyname, "test-executor", colony_prvkey)
    except Exception as err:
        print(err)
        os._exit(0)
    
    print("Executor", "test-executor", "registered")
    return executor_prvkey
    
def unregister(colonies, colonyname, colony_prvkey):
    colonies.remove_executor(colonyname, "test-executor", colony_prvkey)
    print("Executor", "test-executor", "unregistered")

def main():
    colonies, colonyname, colony_prvkey, _, prvkey = colonies_client()
    executor_prvkey = register(colonies, colonyname, colony_prvkey)

    processes = 100
    
    headers = ['nr_processes', 'submit_time', 'assign_time']
    filename = './performance.csv'
    file_exists = os.path.isfile(filename)
    
    with open(filename, 'a', newline='') as file:
        writer = csv.writer(file)
        
        if not file_exists:
            writer.writerow(headers)
     
        submit_time = 0
        assign_time = 0
        start_time = datetime.now()
        #iterations = 10
        iterations = 1
        interval = np.linspace(1, 10000, 100)
        #interval = np.linspace(500000, 1000000, 3)
        print(interval)
        for processes in interval:
            submit_time = 0
            assign_time = 0
            for i in range(iterations):
                 print("iteration: ", i)
                 start_time = time.perf_counter()
                 for _ in range(int(processes)):
                     submit(colonies, colonyname, prvkey)
                 end_time = time.perf_counter()
                 submit_time = submit_time + (end_time-start_time)
             
                 start_time = time.perf_counter()
                 for _ in range(int(processes)):
                     assign(colonies, colonyname, executor_prvkey)
                 end_time = time.perf_counter()
                 assign_time = assign_time + (end_time-start_time)
     
            submit_time = submit_time / iterations
            assign_time = assign_time / iterations
    
            print("processes:", int(processes), "submit_time:", submit_time, "assign_time:", assign_time)
             
            record = [int(processes), submit_time, assign_time]
            writer.writerow(record)
            file.flush()

    unregister(colonies, colonyname, colony_prvkey)

if __name__ == "__main__":
    main()
