# campbot
A bot that camps

```
campsite-watcher/
├── cmd/
│   └── main.go                  // Entry point of the application
├── models/                      // Contains data structures used across the application
│   ├── user.go                  // User related data structures and methods
│   ├── availability.go          // Availability related data structures and methods
│   └── change_notification.go   // Change notification related data structures and methods
├── providers/                   // Holds the interfaces and implementations for interacting with campsite providers
│   ├── provider.go              // Interface definition for campsite providers
│   ├── recreationgov.go         // Recreation.gov provider implementation
│   └── reservecalifornia.go     // ReserveCalifornia provider implementation
├── monitoring/                  // Service for checking campsite availability and finding changes
│   ├── monitor.go               // Methods for monitoring campsite availability
│   └── trigger.go               // Methods for detecting changes and triggering Cloud Function
├── notifications/               // Service for notifying users of changes in campsite availability
│   └── notifier.go              // Methods for sending out notifications
└── database/                    // (Optional, if needed) Service for interacting with the database
    └── db.go                    // Methods for interacting with the database
```

