import pandas as pd
import matplotlib.pyplot as plt

data = pd.read_csv('performance.csv')  # Make sure to use the correct file path

plt.figure(figsize=(10, 5))  # Adjust the size as you like
plt.plot(data['nr_processes'], data['submit_time'], label='Submit Time', marker='o')

plt.plot(data['nr_processes'], data['assign_time'], label='Assign Time', marker='x')

# Adding title and labels
plt.title('Colonies performance')
plt.xlabel('Number of Processes')
plt.ylabel('Time')
plt.legend()

# Show the plot
plt.show()

