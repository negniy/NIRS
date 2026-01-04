from dataclasses import dataclass
import os


@dataclass(frozen=True)
class DetectorConfig:
    weights: str = os.getenv("DETECTOR_WEIGHTS", "yolov8n.pt")
    device: str = os.getenv("DETECTOR_DEVICE", "cpu")
    conf_thres: float = float(os.getenv("DETECTOR_CONF", "0.25"))
    iou_thres: float = float(os.getenv("DETECTOR_IOU", "0.45"))
    imgsz: int = int(os.getenv("DETECTOR_IMGSZ", "640"))
