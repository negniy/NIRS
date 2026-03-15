import argparse
import os
import random
import cv2
import csv
import numpy as np

from app.config import DetectorConfig
from app.detector.yolo import YOLODetector
from app.detector.preprocessing import preprocess_frame
from app.utils.logger import get_logger

log = get_logger("cli_detect")


def sample_frame_indices(total_frames: int, n: int, seed: int) -> list[int]:
    if total_frames <= 0:
        return []
    rng = random.Random(seed)
    n = min(n, total_frames)
    # выбор без повторов
    return sorted(rng.sample(range(total_frames), n))


def draw_boxes(frame: np.ndarray, detections):
    
    out = frame.copy()
    for d in detections:
        x1, y1, x2, y2 = int(d.bbox.x1), int(d.bbox.y1), int(d.bbox.x2), int(d.bbox.y2)
        cv2.rectangle(out, (x1, y1), (x2, y2), (0, 255, 0), 2)
        label = f"cls={d.cls_id} conf={d.conf:.2f}"
        cv2.putText(out, label, (x1, max(0, y1 - 5)), cv2.FONT_HERSHEY_SIMPLEX, 0.5, (0, 255, 0), 1)
    return out


def main():
    ap = argparse.ArgumentParser(description="Run YOLO detection on random frames from a video")
    ap.add_argument("--video", required=True, help="Path to video file")
    ap.add_argument("--weights", default=os.getenv("DETECTOR_WEIGHTS", "yolov8n.pt"))
    ap.add_argument("--device", default=os.getenv("DETECTOR_DEVICE", "cpu"))
    ap.add_argument("--conf", type=float, default=float(os.getenv("DETECTOR_CONF", "0.25")))
    ap.add_argument("--iou", type=float, default=float(os.getenv("DETECTOR_IOU", "0.45")))
    ap.add_argument("--imgsz", type=int, default=int(os.getenv("DETECTOR_IMGSZ", "640")))
    ap.add_argument("--frames", type=int, default=10, help="How many random frames to sample")
    ap.add_argument("--seed", type=int, default=123, help="Random seed for frame sampling")
    ap.add_argument("--out_dir", default="outputs_frames", help="Directory to save annotated frames")
    ap.add_argument("--save", action="store_true", help="Save annotated frames")
    ap.add_argument("--show", action="store_true", help="Show frames in window")

    args = ap.parse_args()
    
    csv_file = open("detections.csv", "w", newline="")
    csv_writer = csv.writer(csv_file)

    csv_writer.writerow([
        "frame",
        "cls",
        "conf",
        "x1",
        "y1",
        "x2",
        "y2"
    ])

    if not os.path.exists(args.video):
        raise FileNotFoundError(args.video)

    cfg = DetectorConfig(
        weights=args.weights,
        device=args.device,
        conf_thres=args.conf,
        iou_thres=args.iou,
        imgsz=args.imgsz,
    )

    detector = YOLODetector(cfg)

    cap = cv2.VideoCapture(args.video)
    if not cap.isOpened():
        raise RuntimeError(f"Cannot open video: {args.video}")

    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    log.info(f"Video opened: frames={total_frames}")

    idxs = sample_frame_indices(total_frames, args.frames, args.seed)
    if not idxs:
        log.warning("No frames sampled (unknown frame count?). Will read sequentially first N frames.")
        idxs = list(range(args.frames))

    if args.save:
        os.makedirs(args.out_dir, exist_ok=True)

    for i, frame_idx in enumerate(idxs):
        cap.set(cv2.CAP_PROP_POS_FRAMES, frame_idx)
        ok, frame = cap.read()
        if not ok or frame is None:
            log.warning(f"Failed to read frame {frame_idx}")
            continue

        frame = preprocess_frame(frame)
        result = detector.detect(frame, frame_index=frame_idx)
        for d in result.detections:
            b = d.bbox
            csv_writer.writerow([
                frame_idx,
                d.cls_id,
                d.conf,
                b.x1,
                b.y1,
                b.x2,
                b.y2
            ])

        # log.info(f"[{i+1}/{len(idxs)}] frame={frame_idx} detections={len(result.detections)}")
        # выводим первые несколько детекций в консоль
        # for d in result.detections[:10]:
        #     b = d.bbox
        #     log.info(f"  cls={d.cls_id} conf={d.conf:.3f} xyxy=({b.x1:.1f},{b.y1:.1f},{b.x2:.1f},{b.y2:.1f})")

        # if args.save or args.show:
        #     annotated = draw_boxes(frame, result.detections)
        #     if args.save:
        #         out_path = os.path.join(args.out_dir, f"frame_{frame_idx:06d}.jpg")
        #         cv2.imwrite(out_path, annotated)
        #     if args.show:
        #         cv2.imshow("detections", annotated)
        #         if cv2.waitKey(1) & 0xFF == ord("q"):
        #             break

    csv_file.close()
    cap.release()
    if args.show:
        cv2.destroyAllWindows()


if __name__ == "__main__":
    main()
