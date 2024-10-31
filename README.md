<div align="center">

# 🐹 GreenGo 💚

</div>

# 🌱 Automated Deployment Pipeline with Docker & Git Integration

Welcome to this deployment automation project, designed to efficiently deploy changes in Git repositories using `docker-compose`. It includes a cycle to review new commits and automatically clone repositories into the runtime environment. This tool facilitates continuous integration and automatic, reliable, and efficient application deployment.

## 💡 Motivation

The original idea behind this project greatly inspired me, prompting a language transition to Go with an emphasis on Docker Compose. This adjustment aligns well with the technologies I frequently use for my university projects, particularly Docker Compose, which remains central to my workflow. Creating this project has significantly streamlined my CI/CD processes, enabling a smoother handling of changes and automations.

## 🌍 Green Coding: Sustainable and Efficient Code

This project was developed following **Green Coding** principles, which aim to minimize the environmental impact of software development. By designing optimized deployment scripts and automating repetitive tasks, we reduce processing time and required computational resources, achieving a more sustainable execution. These small changes, when applied on a large scale, help reduce the carbon footprint of technology projects and build a greener future.

## ⚙️ Setup and Usage

Follow these steps to set up and run this project on your local machine.

1. **Clone this repository:**

    ```bash
    git clone <REPOSITORY_URL>
    cd <REPOSITORY_NAME>
    ```

2. **Configure the required variables:** In the main.go file, update the REPO and BRANCH constants with the repository URL and the branch you want to monitor.

    ```golang
    const (
        REPO   = "https://github.com/your_user/your_repository.git" // Repository URL
        BRANCH = "main" // Branch to monitor
    )
    ```

3. **Build and run the program:** Compile and run the script.

    ```bash
    go build -o deployment-pipeline main.go
    ./deployment-pipeline
    ```

4. **Start the monitoring process:** Once started, the script will check every 5 seconds for new commits on the specified branch and will automatically run docker-compose if it detects changes.

## 🔧 Extending the Pipeline

This project includes a flexible structure that makes it easy to add new stages to the deployment pipeline. Each stage can be defined as a command executed with the helper function `pipelineStage`, located in `pipeline.go`. Here’s how you can add additional steps:

1. **Create a New Stage:**
   To add a new step, simply call the `pipelineStage` function in `processPipeline`, passing the command and working directory as arguments. For example, if you want to add a step to update dependencies, add the following line in `processPipeline`:

    ```go
   pipelineStage([]string{"your-command-here"}, "working-directory")
   ```

2. **Organize Additional Utilities:** For more advanced customizations or reusable functions, consider adding a new file in the utils directory to keep the pipeline organized and modular.

3. **Run Custom Stages Conditionally:** If some stages should run only under certain conditions, you can add conditional statements around pipelineStage calls to control when they execute.

This approach ensures that each pipeline stage is modular and easy to maintain, allowing you to expand functionality without complicating the main workflow.

## 🚀 Credits

This project was developed based on the repository [python-deployment-script-sample](https://github.com/acoronadoc/python-deployment-script-sample) created by [acoronadoc](https://github.com/acoronadoc). From this example, we expanded and optimized the workflow to make it modular, efficient, and easy to integrate into deployments with `docker-compose`.
