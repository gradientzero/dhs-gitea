from dvclive import Live
import random

def run_simple_experiment():

    datapoints = [
        {"name": "petal_width", "importance": 0.4},
        {"name": "petal_length", "importance": 0.33},
        {"name": "sepal_width", "importance": 0.24},
        {"name": "sepal_length", "importance": 0.03}
    ]

    with Live() as live:
        live.log_param("myParam", random.randint(0, 500))
        live.log_metric("myMetric", random.randint(0, 500))
        live.log_metric("new_metric", random.randint(0, 500))

        live.log_plot(
            "iris_feature_importance",
            datapoints,
            x="importance",
            y="name",
            template="bar_horizontal",
            title="Iris Dataset: Feature Importance",
            y_label="Feature Name",
            x_label="Feature Importance"
        )
