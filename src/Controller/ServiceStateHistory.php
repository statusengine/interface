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


use Statusengine\Loader\ServiceStateHistoryLoaderInterface;
use Statusengine\ValueObjects\ServiceStateHistoryQueryOptions;

class ServiceStateHistory extends Controller {

    /**
     * @var ServiceStateHistoryLoaderInterface
     */
    private $ServiceStateHistoryLoader;

    /**
     * ServiceStateHistory constructor.
     * @param ServiceStateHistoryLoaderInterface $ServiceStateHistoryLoader
     */
    public function __construct(ServiceStateHistoryLoaderInterface $ServiceStateHistoryLoader) {
        $this->ServiceStateHistoryLoader = $ServiceStateHistoryLoader;
    }

    /**
     * @param ServiceStateHistoryQueryOptions $queryOptions
     * @return array
     */
    public function index(ServiceStateHistoryQueryOptions $queryOptions) {
        return $this->ServiceStateHistoryLoader->getServiceStateHistory($queryOptions);
    }

}
