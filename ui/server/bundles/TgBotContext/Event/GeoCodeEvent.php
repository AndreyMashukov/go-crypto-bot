<?php
namespace Bundles\TgBotContext\Event;

use App\Context\Geo\Model\Point;
use Symfony\Contracts\EventDispatcher\Event;

class GeoCodeEvent extends Event
{
    private array $results = [];

    private ?Point $point = null;

    private ?string $selectedKind = null;

    public function __construct(private readonly string $toGeocode, private readonly ?int $selectedIndex = null)
    {
    }

    public function getToGeocode(): string
    {
        return $this->toGeocode;
    }

    public function getSelectedIndex(): ?int
    {
        return $this->selectedIndex;
    }

    public function getPoint(): ?Point
    {
        return $this->point;
    }

    public function setPoint(?Point $point): self
    {
        $this->point = $point;

        return $this;
    }

    public function getResults(): array
    {
        return $this->results;
    }

    public function setResults(array $results): self
    {
        $this->results = $results;

        return $this;
    }

    public function getSelectedKind(): ?string
    {
        return $this->selectedKind;
    }

    public function setSelectedKind(?string $selectedKind): self
    {
        $this->selectedKind = $selectedKind;

        return $this;
    }
}
