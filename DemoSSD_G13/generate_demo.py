from pathlib import Path
from PIL import Image, ImageDraw, ImageFont


WIDTH, HEIGHT = 1920, 1080
ROOT = Path(__file__).parent
FRAMES = ROOT / "frames"

BG = "#08111f"
PANEL = "#101d30"
PANEL_2 = "#16263d"
TEXT = "#f4f7fb"
MUTED = "#9eb0c9"
CYAN = "#45d6e8"
GREEN = "#48d597"
ORANGE = "#ffb454"
RED = "#ff6d7a"
PURPLE = "#ad8cff"


def font(size, bold=False):
    names = ["seguisb.ttf" if bold else "segoeui.ttf", "arialbd.ttf" if bold else "arial.ttf"]
    for name in names:
        path = Path("C:/Windows/Fonts") / name
        if path.exists():
            return ImageFont.truetype(str(path), size)
    return ImageFont.load_default()


F_SMALL = font(24)
F_BODY = font(30)
F_BODY_BOLD = font(30, True)
F_TITLE = font(58, True)
F_HERO = font(88, True)
F_LABEL = font(22, True)


def wrap(draw, text, fnt, max_width):
    words = text.split()
    lines, line = [], ""
    for word in words:
        trial = f"{line} {word}".strip()
        if draw.textbbox((0, 0), trial, font=fnt)[2] <= max_width:
            line = trial
        else:
            if line:
                lines.append(line)
            # Las palabras largas también deben respetar el ancho del panel.
            while draw.textbbox((0, 0), word, font=fnt)[2] > max_width:
                cut = len(word) - 1
                while draw.textbbox((0, 0), word[:cut] + "-", font=fnt)[2] > max_width:
                    cut -= 1
                lines.append(word[:cut] + "-")
                word = word[cut:]
            line = word
    if line:
        lines.append(line)
    return lines


def text_block(draw, xy, text, fnt, fill, max_width, gap=10):
    x, y = xy
    for line in wrap(draw, text, fnt, max_width):
        draw.text((x, y), line, font=fnt, fill=fill)
        y += fnt.size + gap
    return y


def fitted(draw, box, value, size, fill, bold=False, min_size=18, gap=7):
    """Ajusta el tamaño y el salto de línea dentro de un rectángulo fijo."""
    left, top, right, bottom = box
    for current in range(size, min_size - 1, -1):
        face = font(current, bold)
        lines = wrap(draw, value, face, right - left)
        line_height = current + gap
        if len(lines) * line_height <= bottom - top:
            for n, line in enumerate(lines):
                draw.text((left, top + n * line_height), line, font=face, fill=fill)
            return
    raise ValueError(f"Texto demasiado largo para el panel: {value}")


def rounded(draw, box, fill, radius=24, outline=None, width=1):
    draw.rounded_rectangle(box, radius=radius, fill=fill, outline=outline, width=width)


def pill(draw, x, y, label, color):
    w = draw.textbbox((0, 0), label, font=F_LABEL)[2] + 30
    rounded(draw, (x, y, x + w, y + 42), color, 20)
    draw.text((x + 15, y + 9), label, font=F_LABEL, fill=BG)
    return w


def icon(draw, kind, cx, cy, color):
    if kind == "map":
        draw.ellipse((cx - 17, cy - 17, cx + 17, cy + 17), fill=color)
        draw.ellipse((cx - 7, cy - 7, cx + 7, cy + 7), fill=BG)
        draw.line((cx, cy + 14, cx, cy + 32), fill=color, width=5)
    elif kind == "bike":
        draw.ellipse((cx - 36, cy - 8, cx - 8, cy + 20), outline=color, width=5)
        draw.ellipse((cx + 10, cy - 8, cx + 38, cy + 20), outline=color, width=5)
        draw.line((cx - 22, cy + 6, cx, cy - 20, cx + 24, cy + 6, cx - 22, cy + 6), fill=color, width=5)
    elif kind == "chart":
        draw.line((cx - 32, cy + 25, cx - 32, cy - 25), fill=color, width=4)
        draw.line((cx - 32, cy + 25, cx + 35, cy + 25), fill=color, width=4)
        for i, h in enumerate((25, 52, 38, 72)):
            x = cx - 18 + i * 17
            draw.rounded_rectangle((x, cy + 20 - h, x + 10, cy + 20), radius=4, fill=color)
    elif kind == "wrench":
        draw.line((cx - 25, cy + 22, cx + 22, cy - 25), fill=color, width=9)
        draw.ellipse((cx - 32, cy - 33, cx - 10, cy - 11), outline=color, width=7)
        draw.ellipse((cx + 13, cy + 14, cx + 32, cy + 33), fill=color)
    elif kind == "alert":
        draw.polygon((cx, cy - 34, cx + 35, cy + 30, cx - 35, cy + 30), fill=color)
        draw.line((cx, cy - 14, cx, cy + 10), fill=BG, width=6)
        draw.ellipse((cx - 3, cy + 17, cx + 3, cy + 23), fill=BG)


def mock_screen(draw, kind, title, tag, accent):
    """Vista conceptual del módulo; todas las etiquetas quedan en sus paneles."""
    x0, y0, x1, y1 = 1000, 272, 1840, 944
    rounded(draw, (x0, y0, x1, y1), PANEL, 24, outline="#34506e", width=2)
    rounded(draw, (x0 + 20, y0 + 20, x1 - 20, y0 + 88), "#0a1729", 14)
    draw.ellipse((x0 + 48, y0 + 47, x0 + 62, y0 + 61), fill=RED)
    draw.ellipse((x0 + 73, y0 + 47, x0 + 87, y0 + 61), fill=ORANGE)
    draw.ellipse((x0 + 98, y0 + 47, x0 + 112, y0 + 61), fill=GREEN)
    fitted(draw, (x0 + 145, y0 + 39, x1 - 190, y0 + 75), "SI-16 · Vista ilustrativa", 27, TEXT, True)
    fitted(draw, (x1 - 140, y0 + 42, x1 - 55, y0 + 73), "DEMO", 20, accent, True)
    fitted(draw, (x0 + 42, y0 + 116, x1 - 42, y0 + 177), title, 45, TEXT, True)
    draw.line((x0 + 42, y0 + 192, x1 - 42, y0 + 192), fill="#334763", width=2)

    if kind == "map":
        rounded(draw, (x0 + 42, y0 + 220, x1 - 42, y1 - 105), "#18334a", 18)
        for i in range(7):
            px = x0 + 135 + (i * 91) % 610
            py = y0 + 290 + (i * 77) % 225
            draw.line((px - 55, py + 20, px + 58, py - 28), fill="#2d6681", width=6)
            icon(draw, "map", px, py, GREEN if i % 3 else ORANGE)
        pill(draw, x0 + 45, y1 - 80, "60 estaciones", CYAN)
        pill(draw, x0 + 240, y1 - 80, "Bicis y anclajes", GREEN)
    elif kind == "bike":
        icon(draw, "bike", x0 + 108, y0 + 285, accent)
        if "caducado" in title.lower():
            state, detail, badge = "Desbloqueo rechazado", "El abono debe estar vigente", "Sin cambios en el anclaje"
        elif "segundo" in title.lower():
            state, detail, badge = "Viaje ya activo", "No puede iniciar otro trayecto", "Sin segundo alquiler"
        elif "gratuito" in title.lower():
            state, detail, badge = "Viaje finalizado", "Duración ≤ 30 min · 0,00 €", "Tramo gratuito"
        elif "exceso" in title.lower():
            state, detail, badge = "Viaje finalizado", "Se cobra solo el exceso", "Tarifa calculada"
        elif "finalizar" in title.lower():
            state, detail, badge = "Bicicleta anclada", "Viaje finalizado correctamente", "Anclaje actualizado"
        else:
            state, detail, badge = "Desbloqueo correcto", "Bicicleta libre · viaje activo", "Viaje iniciado"
        fitted(draw, (x0 + 175, y0 + 255, x1 - 55, y0 + 325), state, 37, TEXT, True)
        rounded(draw, (x0 + 42, y0 + 355, x1 - 42, y0 + 515), PANEL_2, 18)
        fitted(draw, (x0 + 70, y0 + 380, x1 - 70, y0 + 430), "Bicicleta y anclaje", 31, TEXT, True)
        fitted(draw, (x0 + 70, y0 + 450, x1 - 70, y0 + 495), detail, 27, MUTED)
        pill(draw, x0 + 48, y1 - 85, badge, accent)
    elif kind == "history":
        for i, value in enumerate(("15 oct  ·  18 min  ·  0,00 €", "08 oct  ·  42 min  ·  1,20 €", "02 oct  ·  25 min  ·  0,00 €")):
            yy = y0 + 235 + i * 103
            rounded(draw, (x0 + 42, yy, x1 - 42, yy + 75), PANEL_2, 13)
            fitted(draw, (x0 + 70, yy + 18, x1 - 65, yy + 59), value, 29, TEXT)
        pill(draw, x0 + 45, y1 - 80, "Filtrar mes y año", PURPLE)
    elif kind == "wrench":
        if "reportar" in title.lower():
            heading, summary, field, detail, badge = "Nuevo reporte", "Bicicleta B-024 · estación", "Descripción de la avería", "Freno delantero defectuoso", "Enviar reporte"
        elif "inhabilitación" in title.lower():
            heading, summary, field, detail, badge = "Bicicleta averiada", "Bicicleta B-024 · estación", "Estado de alquiler", "No disponible para nuevos viajes", "Retirada del conteo"
        elif "aviso" in title.lower():
            heading, summary, field, detail, badge = "Incidencia pendiente", "Bicicleta B-024 · freno", "Cola del taller", "El operario recibe el aviso", "Pendiente de revisión"
        elif "ficha" in title.lower() or "parte" in title.lower():
            heading, summary, field, detail, badge = "Ficha de bicicleta", "Bicicleta B-024 · averiada", "Reporte original", "Freno delantero defectuoso", "Historial visible"
        elif "registrar" in title.lower():
            heading, summary, field, detail, badge = "Intervención técnica", "Bicicleta B-024 · averiada", "Parte de reparación", "Piezas · notas · operario", "Guardar intervención"
        else:
            heading, summary, field, detail, badge = "Bicicleta operativa", "Bicicleta B-024 · reparada", "Confirmación de anclaje", "De nuevo visible en el mapa", "Disponible"
        fitted(draw, (x0 + 42, y0 + 198, x1 - 42, y0 + 252), heading, 36, TEXT, True)
        rounded(draw, (x0 + 42, y0 + 260, x1 - 42, y0 + 360), PANEL_2, 14)
        fitted(draw, (x0 + 70, y0 + 282, x1 - 70, y0 + 340), summary, 29, TEXT, True)
        rounded(draw, (x0 + 42, y0 + 380, x1 - 42, y0 + 530), PANEL_2, 14)
        fitted(draw, (x0 + 70, y0 + 403, x1 - 70, y0 + 453), field, 31, accent, True)
        fitted(draw, (x0 + 70, y0 + 463, x1 - 70, y0 + 510), detail, 27, MUTED)
        pill(draw, x0 + 45, y1 - 80, badge, accent)
    elif kind == "alert":
        for i, (station, value, color) in enumerate((("Plaça Ajuntament", "95 %", RED), ("Estació Nord", "5 %", ORANGE), ("Mercat Central", "92 %", RED))):
            yy = y0 + 235 + i * 100
            rounded(draw, (x0 + 42, yy, x1 - 42, yy + 75), PANEL_2, 14)
            draw.ellipse((x0 + 70, yy + 25, x0 + 92, yy + 47), fill=color)
            fitted(draw, (x0 + 112, yy + 18, x1 - 170, yy + 58), station, 29, TEXT, True)
            fitted(draw, (x1 - 145, yy + 18, x1 - 60, yy + 58), value, 27, color, True)
        pill(draw, x0 + 45, y1 - 80, "Prioridad alta", accent)
    elif kind == "chart":
        rounded(draw, (x0 + 42, y0 + 230, x1 - 42, y0 + 520), PANEL_2, 14)
        if "vacío" not in title.lower():
            for i, height in enumerate((92, 158, 120, 204, 145, 228, 175)):
                xx = x0 + 85 + i * 92
                draw.rounded_rectangle((xx, y0 + 475 - height, xx + 49, y0 + 475), radius=8, fill=CYAN if i % 2 else PURPLE)
        draw.line((x0 + 72, y0 + 476, x1 - 75, y0 + 476), fill="#59718d", width=2)
        pill(draw, x0 + 45, y1 - 80, "Por hora · por día", CYAN)
        pill(draw, x0 + 270, y1 - 80, "Exportar CSV", GREEN)


def make_scene(index, total, story, title, actor, action, result, kind, accent, tag):
    image = Image.new("RGB", (WIDTH, HEIGHT), BG)
    draw = ImageDraw.Draw(image)
    draw.rectangle((0, 0, WIDTH, 10), fill=accent)
    draw.ellipse((1410, -210, 2060, 350), outline="#17314b", width=2)
    draw.ellipse((1480, -140, 1980, 280), outline="#17314b", width=2)
    rounded(draw, (80, 65, 195, 169), accent, 21)
    fitted(draw, (97, 93, 180, 147), story.split(" · ")[0], 34, BG, True)
    fitted(draw, (230, 69, 1640, 167), story.split(" · ", 1)[1], 59, TEXT, True)
    draw.line((80, 200, 1840, 200), fill="#31465e", width=2)
    pill(draw, 80, 225, tag, accent)
    fitted(draw, (80, 294, 925, 435), title, 67, TEXT, True, min_size=36, gap=5)
    draw.text((80, 479), "ACTOR", font=F_LABEL, fill=accent)
    fitted(draw, (80, 515, 925, 575), actor, 33, TEXT, True)
    draw.text((80, 608), "ACCIÓN", font=F_LABEL, fill=accent)
    fitted(draw, (80, 647, 925, 758), action, 34, MUTED, min_size=26)
    rounded(draw, (80, 796, 925, 945), PANEL_2, 20, outline="#34506e", width=2)
    draw.rectangle((80, 812, 86, 928), fill=accent)
    draw.text((110, 818), "RESULTADO", font=F_LABEL, fill=accent)
    fitted(draw, (110, 860, 892, 932), result, 31, TEXT, True, min_size=22, gap=3)
    mock_screen(draw, kind, title, tag, accent)
    draw.text((80, 1002), f"ESCENA {index:02d}/{total:02d}", font=F_SMALL, fill=MUTED)
    draw.text((560, 1002), "SI-16  ·  HISTORIAS DE USUARIO", font=F_SMALL, fill=MUTED)
    draw.rounded_rectangle((80, 1042, 1840, 1052), radius=5, fill="#20334d")
    draw.rounded_rectangle((80, 1042, 80 + int(1760 * index / total), 1052), radius=5, fill=accent)
    return image


def make_cover(closing=False):
    """Portada y cierre del vídeo, con la misma identidad visual."""
    image = Image.new("RGB", (WIDTH, HEIGHT), BG)
    draw = ImageDraw.Draw(image)
    draw.rectangle((0, 0, WIDTH, 12), fill=CYAN)
    draw.ellipse((1200, -300, 2200, 700), outline="#1b3854", width=4)
    draw.ellipse((1300, -200, 2100, 600), outline="#1b3854", width=4)
    pill(draw, 92, 105, "SI-16 · BICICLETAS COMPARTIDAS", CYAN)
    if closing:
        fitted(draw, (92, 250, 1710, 390), "Fin de la demostración", 90, TEXT, True)
        fitted(draw, (92, 420, 1610, 545), "8 historias de usuario · 24 escenarios ilustrados", 49, MUTED)
        draw.text((92, 745), "Reproducción por historia en index.html", font=font(36, True), fill=CYAN)
    else:
        fitted(draw, (92, 250, 1710, 390), "La red en movimiento", 102, TEXT, True)
        fitted(draw, (92, 420, 1610, 555), "Una visita guiada a las ocho historias de usuario", 53, MUTED)
        draw.text((92, 745), "USUARIOS  /  TALLER  /  LOGÍSTICA  /  GESTIÓN", font=font(32, True), fill=CYAN)
    draw.line((92, 905, 1828, 905), fill="#34506e", width=3)
    draw.text((92, 953), "DEMO SSD  ·  SI-16", font=font(28, True), fill=MUTED)
    draw.text((1580, 953), "2026", font=font(28, True), fill=MUTED)
    return image


def main():
    scenes = [
        ("HU-01 · Disponibilidad en tiempo real", "Mapa de la red", "Usuario con abono activo", "Abrir la pantalla principal y consultar las estaciones de València.", "El mapa muestra unas 60 estaciones y sus bicicletas disponibles.", "map", CYAN, "RECORRIDO HABITUAL"),
        ("HU-01 · Disponibilidad en tiempo real", "Detalle de una estación", "Usuario registrado", "Seleccionar una estación concreta en el mapa.", "Se ven bicicletas disponibles y anclajes libres con precisión.", "map", GREEN, "DETALLE"),
        ("HU-01 · Disponibilidad en tiempo real", "Actualización en vivo", "Usuario esperando una bicicleta", "Cambiar el estado de una estación y esperar la actualización.", "La vista refleja el nuevo estado sin superar el retardo previsto.", "map", ORANGE, "ESTADO / LÍMITE"),
        ("HU-02 · Desbloqueo", "Iniciar un viaje", "Usuario con abono vigente", "Introducir el código de un anclaje con bicicleta disponible.", "La bicicleta se libera y se crea un viaje activo.", "bike", CYAN, "RECORRIDO HABITUAL"),
        ("HU-02 · Desbloqueo", "Abono caducado", "Usuario con abono inactivo", "Intentar desbloquear una bicicleta con la suscripción caducada.", "La operación se rechaza y se informa de la causa.", "bike", RED, "PERMISO / ERROR"),
        ("HU-02 · Desbloqueo", "Segundo viaje bloqueado", "Usuario con viaje activo", "Volver a solicitar otro desbloqueo durante un viaje.", "El sistema impide mantener dos viajes simultáneos.", "bike", ORANGE, "ESTADO INCOMPATIBLE"),
        ("HU-03 · Devolución y tarifa", "Finalizar el viaje", "Usuario en viaje", "Elegir una estación con anclaje libre y fijar la bicicleta.", "Se registra la llegada, se bloquea la bicicleta y se libera el usuario.", "bike", GREEN, "RECORRIDO HABITUAL"),
        ("HU-03 · Devolución y tarifa", "Tramo gratuito", "Usuario en viaje de 30 minutos", "Devolver una bicicleta con duración menor o igual a 30 minutos.", "El coste aplicado es 0,00 € y el viaje queda finalizado.", "bike", CYAN, "CASO LÍMITE"),
        ("HU-03 · Devolución y tarifa", "Exceso de tiempo", "Usuario en viaje largo", "Devolver una bicicleta después de los 30 minutos incluidos.", "Se cobra únicamente el tramo que supera los 30 minutos.", "bike", ORANGE, "TARIFICACIÓN"),
        ("HU-04 · Historial", "Consultar viajes", "Usuario del servicio", "Entrar en la sección Mis Viajes.", "Aparece una lista cronológica con origen, destino, duración y coste.", "history", CYAN, "RECORRIDO HABITUAL"),
        ("HU-04 · Historial", "Filtrar por mes", "Usuario revisando gastos", "Seleccionar un mes y un año concretos.", "Solo se muestran los viajes del rango y su sumatorio.", "history", PURPLE, "FILTRO"),
        ("HU-04 · Historial", "Identificar tramo gratuito", "Usuario con viajes sin cargo", "Revisar un viaje de menos de 30 minutos.", "El historial muestra claramente Tramo gratuito (0,00 €).", "history", GREEN, "CASO LÍMITE"),
        ("HU-05 · Averías", "Reportar una avería", "Usuario que detecta un fallo", "Seleccionar bicicleta, estación y describir un problema de frenos.", "La bicicleta cambia a Averiada y el reporte queda registrado.", "wrench", RED, "RECORRIDO HABITUAL"),
        ("HU-05 · Averías", "Inhabilitación preventiva", "Red de estaciones", "Actualizar el mapa después de marcar la bicicleta como averiada.", "La bicicleta deja de alquilarse y no cuenta como operativa.", "wrench", ORANGE, "ESTADO"),
        ("HU-05 · Averías", "Aviso al taller", "Operario de mantenimiento", "Abrir el panel de incidencias después del reporte.", "La incidencia aparece como pendiente para el equipo técnico.", "wrench", PURPLE, "EVENTO"),
        ("HU-06 · Taller", "Ver el parte", "Operario de taller", "Abrir la ficha de una bicicleta averiada.", "Se muestra el reporte original enviado por el usuario.", "wrench", CYAN, "RECORRIDO HABITUAL"),
        ("HU-06 · Taller", "Registrar reparación", "Operario autenticado", "Añadir piezas sustituidas, notas y confirmar la intervención.", "La reparación queda firmada y guardada en PostgreSQL.", "wrench", GREEN, "FORMULARIO"),
        ("HU-06 · Taller", "Reactivar bicicleta", "Operario de taller", "Marcar la bicicleta como Operativa y confirmar su anclaje.", "Vuelve a aparecer como disponible en el mapa.", "wrench", ORANGE, "CAMBIO DE ESTADO"),
        ("HU-07 · Balanceo", "Detectar estaciones críticas", "Operario de transporte", "Abrir el panel con estaciones por debajo del 10 % o por encima del 90 %.", "Se generan alertas automáticas de prioridad alta.", "alert", RED, "RECORRIDO HABITUAL"),
        ("HU-07 · Balanceo", "Ordenar por severidad", "Operario de calle", "Consultar la lista de balanceo.", "Las estaciones aparecen ordenadas por urgencia de intervención.", "alert", ORANGE, "PRIORIZACIÓN"),
        ("HU-07 · Balanceo", "Mover bicicletas", "Operario con furgoneta", "Confirmar la carga en una estación y la descarga en otra.", "La disponibilidad de ambas estaciones se recalcula de inmediato.", "alert", GREEN, "OPERACIÓN"),
        ("HU-08 · Analítica", "Consultar gráficos", "Gestor municipal", "Establecer un rango de fechas en el panel de analítica.", "Se generan gráficos por hora y por día de la semana.", "chart", CYAN, "RECORRIDO HABITUAL"),
        ("HU-08 · Analítica", "Exportar datos", "Gestor municipal", "Pulsar Descargar CSV para el mismo rango analizado.", "El CSV contiene el histórico filtrado y conserva sus totales.", "chart", GREEN, "EXPORTACIÓN"),
        ("HU-08 · Analítica", "Rango vacío y permisos", "Gestor / usuario sin permisos", "Probar un rango futuro y después acceder con un rol usuario.", "El rango vacío muestra ceros; el rol no gestor recibe 403.", "chart", RED, "LÍMITE / PERMISO"),
    ]
    FRAMES.mkdir(parents=True, exist_ok=True)
    total = len(scenes)
    make_cover().save(FRAMES / "scene-00.png")
    for number, scene in enumerate(scenes, 1):
        make_scene(number, total, *scene).save(FRAMES / f"scene-{number:02d}.png")
    make_cover(closing=True).save(FRAMES / f"scene-{total + 1:02d}.png")
    print(f"Generadas {total} escenas y portada/cierre en {FRAMES}")


if __name__ == "__main__":
    main()
