# Copyright 2025 Google.
#
# This software is provided as-is, without warranty or representation for any use or purpose.
import os
from typing import Dict

from axlearn.experiments.text.gpt import c4_trainer
from axlearn.experiments.trainer_config_utils import TrainerConfigFn
from axlearn.common.trainer import SpmdTrainer
from axlearn.experiments.text.gpt.common import mesh_shape_from_axes
from axlearn.common.config import config_for_function
from axlearn.common.evaler import every_n_steps_policy as eval_every_n_steps_policy
from axlearn.common.checkpointer import every_n_steps_policy

default_eval_every_n_steps = 2_000
default_save_every_n_steps = 2_000
default_keep_every_n_steps = 10_000
default_max_step = 50_000 

def create_config(model: str, batch_size: int, config_map: Dict[str, TrainerConfigFn], 
                  mesh_shape, 
                  max_step: int = default_max_step, 
                  eval_every_n_steps: int = default_eval_every_n_steps,
                  save_every_n_steps: int = default_save_every_n_steps,
                  keep_every_n_steps: int = default_keep_every_n_steps):
    cfg: SpmdTrainer.Config = config_map[model]().clone()
    cfg.input.batcher.global_batch_size = batch_size
    cfg.max_step = max_step
    if mesh_shape is not None:
        cfg.mesh_shape = mesh_shape
    for evaler in cfg.evalers.values():
        evaler.input.batcher.global_batch_size = batch_size
        evaler.set(
                eval_policy=config_for_function(eval_every_n_steps_policy).set(n=eval_every_n_steps)
            )
        
    
    if os.getenv("SAVE_EVERY_N_STEPS", None):
        save_every_n_steps = int(os.getenv("SAVE_EVERY_N_STEPS"))
      
    cfg.checkpointer.save_policy = config_for_function(every_n_steps_policy).set(
            n=save_every_n_steps or min(eval_every_n_steps, 5_000)
        )
    cfg.checkpointer.keep_every_n_steps = min(max_step, keep_every_n_steps)
    return lambda: cfg

def named_trainer_configs() -> Dict[str, TrainerConfigFn]:
    config_map = c4_trainer.named_trainer_configs()
    
    # For single host v4-8. use fuji-7B-s1-b4 size to fit HBM 
    for slice_num in [1, 2, 4]:
        # set data sharding per slice. set fsdp(model sharding) to be infered, which is 
        # number of devices / data sharding
        mesh_shape = mesh_shape_from_axes(data=slice_num, fsdp=-1)
        for batch_size in [4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096]:
            config_map[f"fuji-7B-s{slice_num}-b{batch_size}"] = create_config("fuji-7B-v1", batch_size, config_map, mesh_shape)

             
    return config_map