import csv
import tempfile
import unittest
from pathlib import Path

import vinyl_quoter


class WorkSelectionTests(unittest.TestCase):
    def test_collect_single_image_uses_requested_file(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            src = root / "src"
            src.mkdir()
            image = src / "DSC01.jpg"
            image.write_bytes(b"jpg")
            (src / "DSC02.jpg").write_bytes(b"jpg")

            items = vinyl_quoter.collect_work_items(src, image=image, all_images=False)

            self.assertEqual(items, [image])

    def test_collect_all_images_uses_supported_files_sorted(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            src = root / "src"
            src.mkdir()
            (src / "b.png").write_bytes(b"png")
            (src / "a.jpg").write_bytes(b"jpg")
            (src / "notes.txt").write_text("ignore", encoding="utf-8")

            items = vinyl_quoter.collect_work_items(src, image=None, all_images=True)

            self.assertEqual([item.name for item in items], ["a.jpg", "b.png"])


class CatalogPolicyTests(unittest.TestCase):
    def test_update_mode_keeps_existing_rows_and_skips_processed_images(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            report = root / "report" / "album_catalog.csv"
            report.parent.mkdir()
            existing = vinyl_quoter.AlbumCatalogRow(
                source_image="src/a.jpg",
                artist="Artist A",
                title="Title A",
                identification_confidence="high",
                recommended_price_eur="12",
                price_confidence="medium",
                price_basis="existing",
                notes="",
            )
            vinyl_quoter.write_catalog_csv([existing], report)

            pending = vinyl_quoter.pending_images([Path("src/a.jpg"), Path("src/b.jpg")], report, replace=False)

            self.assertEqual(pending, [Path("src/b.jpg")])

    def test_replace_mode_processes_every_image_even_if_csv_exists(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            report = root / "report" / "album_catalog.csv"
            report.parent.mkdir()
            existing = vinyl_quoter.AlbumCatalogRow(
                source_image="src/a.jpg",
                artist="Artist A",
                title="Title A",
                identification_confidence="high",
                recommended_price_eur="12",
                price_confidence="medium",
                price_basis="existing",
                notes="",
            )
            vinyl_quoter.write_catalog_csv([existing], report)

            pending = vinyl_quoter.pending_images([Path("src/a.jpg"), Path("src/b.jpg")], report, replace=True)

            self.assertEqual(pending, [Path("src/a.jpg"), Path("src/b.jpg")])


class CatalogOutputTests(unittest.TestCase):
    def test_process_images_appends_identified_rows_to_default_report(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            src = root / "src"
            src.mkdir()
            image = src / "DSC01.jpg"
            image.write_bytes(b"jpg")
            report = root / "report" / "album_catalog.csv"

            def identify(path):
                return vinyl_quoter.AlbumCatalogData(
                    artist="The Cure",
                    title="Disintegration",
                    identification_confidence="medium",
                    recommended_price_eur="22",
                    price_confidence="low",
                    price_basis="Spain/EU; VG+/VG; album-level estimate",
                    notes=f"identified from {path.name}",
                )

            rows = vinyl_quoter.process_images([image], report, replace=False, identify=identify)

            self.assertEqual(len(rows), 1)
            self.assertEqual(rows[0].artist, "The Cure")
            with report.open(newline="", encoding="utf-8") as handle:
                csv_rows = list(csv.DictReader(handle))
            self.assertEqual(csv_rows[0]["source_image"], str(image))
            self.assertEqual(csv_rows[0]["recommended_price_eur"], "22")


class CliMenuTests(unittest.TestCase):
    def test_menu_choice_four_means_process_all_and_replace(self):
        config = vinyl_quoter.menu_config(lambda prompt: "4")

        self.assertTrue(config.all_images)
        self.assertTrue(config.replace)
        self.assertIsNone(config.image)

    def test_default_provider_uses_recommended_local_vision_model(self):
        config = vinyl_quoter.RunConfig(image=None, all_images=True, replace=False)

        self.assertEqual(config.provider, "lm-studio")
        self.assertEqual(config.model, vinyl_quoter.DEFAULT_LM_STUDIO_VISION_MODEL)

    def test_menu_can_select_gemini_provider(self):
        responses = iter(["2", "2"])

        config = vinyl_quoter.menu_config(lambda prompt: next(responses))

        self.assertTrue(config.all_images)
        self.assertEqual(config.provider, "gemini")
        self.assertEqual(config.model, vinyl_quoter.DEFAULT_GEMINI_MODEL)

    def test_menu_default_provider_is_local_vision_model(self):
        responses = iter(["2", ""])

        config = vinyl_quoter.menu_config(lambda prompt: next(responses))

        self.assertEqual(config.provider, "lm-studio")
        self.assertEqual(config.model, vinyl_quoter.DEFAULT_LM_STUDIO_VISION_MODEL)


if __name__ == "__main__":
    unittest.main()
