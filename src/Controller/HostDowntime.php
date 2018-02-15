<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Controller;


use Statusengine\Loader\HostDowntimeLoaderInterface;
use Statusengine\Loader\ServiceDowntimeLoaderInterface;
use Statusengine\ValueObjects\HostDowntimeQueryOptions;

class HostDowntime extends Controller {

    /**
     * @var HostDowntimeLoaderInterface
     */
    private $HostDowntimeLoader;

    /**
     * HostDowntime constructor.
     * @param HostDowntimeLoaderInterface $HostDowntimeLoader
     */
    public function __construct(HostDowntimeLoaderInterface $HostDowntimeLoader) {
        $this->HostDowntimeLoader = $HostDowntimeLoader;
    }

    /**
     * @param HostDowntimeQueryOptions $queryOptions
     * @return array
     */
    public function index(HostDowntimeQueryOptions $queryOptions) {
        return $this->HostDowntimeLoader->getHostdowntime($queryOptions);
    }

}
