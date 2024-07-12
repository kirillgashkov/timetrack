# TimeTrack

TimeTrack is a simple time tracking tool that allows you to track your time spent on different tasks.

It is a multi-user server application that stores the time tracking data in a PostgreSQL database and communicates with
(pseudo-) external services to provide additional information about the users.

It was created as a test assignment for a job application in July 2024.

## Features

- **Track time**

  Start and stop timers for tasks <mark>quickly</mark> and <mark>safely</mark>, even with many users at the same time.

- **Generate reports**

  Generate reports for the time spent on tasks in a specific time frame.

- **Manage tasks**

  Create, view, update, and delete tasks.

- **Manage users**

  Register users using their national ID number, view all users with <mark>filtering</mark> options and
  <mark>pagination</mark>, and update or delete user information as needed.

  When registering a user, the application will try to fetch additional information about the user from a (pseudo-)
  <mark>external service</mark>.

  The application uses <mark>token-based authentication</mark> to ensure that only registered users can access the
  application. For simplicity, the application uses user IDs as tokens.
