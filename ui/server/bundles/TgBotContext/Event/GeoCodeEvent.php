<?php
/**
 * This file is private property of the author, keep it secure and do not share anywhere out of the author.
 */

namespace Bundles\TgBotContext\Event;

use App\Context\Geo\Model\Point;
use Symfony\Contracts\EventDispatcher\Event;

class GeoCodeEvent extends Event
{
    private string $toGeocode;

    private array $results = [];

    private ?int $selectedIndex;

    private ?Point $point = null;

    private ?string $selectedKind = null;

    public function __construct(string $toGeocode, int $selectedIndex = null)
    {
        $this->toGeocode     = $toGeocode;
        $this->selectedIndex = $selectedIndex;
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
