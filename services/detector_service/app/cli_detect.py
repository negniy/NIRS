import argparse
import os
import csv
import cv2

from app.config import DetectorConfig
from app.detector.yolo import YOLODetector
from app.detector.preprocessing import preprocess_frame
from app.utils.logger import get_logger

log = get_logger("cli_detect_images")


def main():
    ap = argparse.ArgumentParser(description="Run YOLO detection on images from directory")

    ap.add_argument("--images_dir", required=True, help="Path to images directory")
    ap.add_argument("--weights", default=os.getenv("DETECTOR_WEIGHTS", "yolov8n.pt"))
    ap.add_argument("--device", default=os.getenv("DETECTOR_DEVICE", "cpu"))
    ap.add_argument("--conf", type=float, default=float(os.getenv("DETECTOR_CONF", "0.25")))
    ap.add_argument("--iou", type=float, default=float(os.getenv("DETECTOR_IOU", "0.45")))
    ap.add_argument("--imgsz", type=int, default=int(os.getenv("DETECTOR_IMGSZ", "640")))

    args = ap.parse_args()

    # CSV
    csv_file = open("detections2.csv", "w", newline="")
    csv_writer = csv.writer(csv_file)

    csv_writer.writerow([
        "image",
        "cls",
        "conf",
        "x1",
        "y1",
        "x2",
        "y2"
    ])

    if not os.path.exists(args.images_dir):
        raise FileNotFoundError(args.images_dir)

    # detector
    cfg = DetectorConfig(
        weights=args.weights,
        device=args.device,
        conf_thres=args.conf,
        iou_thres=args.iou,
        imgsz=args.imgsz,
    )

    detector = YOLODetector(cfg)

    # список изображений
    image_files = sorted([
        f for f in os.listdir(args.images_dir)
        if f.lower().endswith((".jpg", ".jpeg", ".png"))
    ])

    log.info(f"Found {len(image_files)} images")

    # основной цикл
    for i, img_name in enumerate(image_files):

        img_path = os.path.join(args.images_dir, img_name)

        frame = cv2.imread(img_path)

        if frame is None:
            log.warning(f"Failed to read {img_name}")
            continue

        frame = preprocess_frame(frame)

        result = detector.detect(frame, frame_index=i)

        for d in result.detections:
            b = d.bbox

            csv_writer.writerow([
                img_name,
                d.cls_id,
                d.conf,
                b.x1,
                b.y1,
                b.x2,
                b.y2
            ])

        log.info(f"[{i+1}/{len(image_files)}] {img_name} detections={len(result.detections)}")

    csv_file.close()

    log.info("Done. Results saved to detections2.csv")


if __name__ == "__main__":
    main()