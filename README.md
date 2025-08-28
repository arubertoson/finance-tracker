# Personal Finance Tracker

A command-line tool for tracking personal finances, analyzing spending patterns, and managing budgets. This tool focuses on understanding your financial habits rather than real-time balance tracking.

## Overview

Personal Finance Tracker helps you understand your spending patterns by categorizing transactions, tracking budgets, and generating insightful reports. It's designed for users who want to analyze their financial habits without the complexity of full-fledged accounting software.

## Key Features

- **Transaction Management**
  - Automatic categorization based on party rules
  - Support for recurring transaction tracking
  - Flexible transaction categorization

- **Budgeting**
  - Monthly budget planning
  - Budget vs. actual spending analysis
  - Savings goals tracking

- **Comprehensive Reporting**
  - Monthly spending summaries
  - Category-based analysis
  - Top spending patterns
  - Trend analysis
  - Custom date range reports

- **Data Export**
  - Support for multiple formats (CSV, JSON, Markdown)
  - Customizable report outputs

- **Smart Data Import**
  - Automatic header detection and mapping
  - Support for multiple file formats (CSV, JSON)
  - Custom field mapping configurations
  - Intelligent date format detection
  - Batch import capabilities

## Tech Stack

- **Language:** Go
- **Database:** SQLite
- **Key Dependencies:**
  - `github.com/urfave/cli/v2` - CLI framework
  - `github.com/mattn/go-sqlite3` - SQLite driver
  - Standard Go libraries

## Getting Started

### Prerequisites

- Go 1.21 or higher
- SQLite 3
- Git (for installation)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/finance-tracker.git
cd finance-tracker
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o finance
```

4. Initialize the database:
```bash
./finance init
```

## Documentation

Detailed documentation for each feature can be found in the `docs` directory:

- [Transaction Management](docs/transactions.md)
- [Budget Management](docs/budgets.md)
- [Import/Export](docs/data-management.md)
- [Reports](docs/reports.md)
- [Bank Templates](docs/bank-templates.md)
- [Configuration](docs/configuration.md)

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.