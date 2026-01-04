from pydantic import BaseModel
from typing import List


class BBoxXYXY(BaseModel):
    x1: float
    y1: float
    x2: float
    y2: float


class Detection(BaseModel):
    cls_id: int
    conf: float
    bbox: BBoxXYXY


class DetectionResult(BaseModel):
    frame_index: int
    width: int
    height: int
    detections: List[Detection]
