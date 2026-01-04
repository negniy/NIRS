from ultralytics import YOLO
import numpy as np
from typing import List, Tuple

from app.schemas.detection import Detection, DetectionResult, BBoxXYXY
from app.config import DetectorConfig
from app.utils.logger import get_logger


log = get_logger("yolo")


class YOLODetector:
    """
    Обертка над Ultralytics YOLOv8.
    Возвращает bbox в xyxy + conf + cls_id.
    """

    def __init__(self, cfg: DetectorConfig):
        self.cfg = cfg
        log.info(f"Loading YOLO weights={cfg.weights} device={cfg.device}")
        self.model = YOLO(cfg.weights)

    def detect(self, frame_bgr: np.ndarray, frame_index: int = -1) -> DetectionResult:
        h, w = frame_bgr.shape[:2]
        # Ultralytics принимает numpy BGR нормально
        results = self.model.predict(
            source=frame_bgr,
            conf=self.cfg.conf_thres,
            iou=self.cfg.iou_thres,
            imgsz=self.cfg.imgsz,
            device=self.cfg.device,
            verbose=False,
        )

        dets: List[Detection] = []
        r0 = results[0]
        if r0.boxes is not None and len(r0.boxes) > 0:
            boxes_xyxy = r0.boxes.xyxy.cpu().numpy()
            confs = r0.boxes.conf.cpu().numpy()
            clss = r0.boxes.cls.cpu().numpy().astype(int)

            for (x1, y1, x2, y2), conf, cls_id in zip(boxes_xyxy, confs, clss):
                dets.append(
                    Detection(
                        cls_id=int(cls_id),
                        conf=float(conf),
                        bbox=BBoxXYXY(x1=float(x1), y1=float(y1), x2=float(x2), y2=float(y2)),
                    )
                )

        return DetectionResult(frame_index=frame_index, width=w, height=h, detections=dets)
