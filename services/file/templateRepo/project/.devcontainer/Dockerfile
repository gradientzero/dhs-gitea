FROM python:3.13

## OS dependencies: dvc exp show reuires 'less'
RUN apt update && apt install -y less

# Add non-root user
ARG USERNAME=nonroot
RUN groupadd --gid 1000 $USERNAME && \
    useradd --uid 1000 --gid 1000 -m $USERNAME

## Make sure to reflect new user in PATH
ENV PATH="/home/${USERNAME}/.local/bin:${PATH}"
USER $USERNAME

## Pip dependencies
# Upgrade pip
RUN pip install --upgrade pip

# Install production dependencies
COPY --chown=nonroot:1000 requirements.txt /tmp/requirements.txt
RUN pip install -r /tmp/requirements.txt && \
    rm /tmp/requirements.txt

# Install development dependencies
COPY --chown=nonroot:1000 requirements-dev.txt /tmp/requirements-dev.txt
RUN pip install -r /tmp/requirements-dev.txt && \
    rm /tmp/requirements-dev.txt

# fix: https://github.com/iterative/dvc/issues/10431
#RUN pip install pygit2==1.14.1

# copy all over
# COPY . .
